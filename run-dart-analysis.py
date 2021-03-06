#!/usr/bin/env python
# Copyright 2016 The Fuchsia Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import argparse
import json
import multiprocessing
import os
import paths
import Queue
import subprocess
import sys
import threading


def gn_describe(out, path):
    gn = os.path.join(paths.FUCHSIA_ROOT, 'packages', 'gn', 'gn.py')
    data = subprocess.check_output(
        [gn, 'desc', out, path, '--format=json'], cwd=paths.FUCHSIA_ROOT)
    return json.loads(data)


class WorkerThread(threading.Thread):
    '''
    A worker thread to run scripts from a queue and return exit codes and output on a queue.
    '''

    def __init__(self, script_queue, result_queue, args):
        threading.Thread.__init__(self)
        self.script_queue = script_queue
        self.result_queue = result_queue
        self.args = args

    def run(self):
        while True:
            try:
                script = self.script_queue.get(False)
            except Queue.Empty, e:
                # no more scripts to run
                return
            job = subprocess.Popen(
                [script] + self.args,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE)
            stdout, stderr = job.communicate()
            self.result_queue.put((script, job.returncode, stdout + stderr))


def main():
    parser = argparse.ArgumentParser(
        '''Run Dart analysis for Dart build targets
Extra flags will be passed to the analyzer.
''')
    parser.add_argument(
        '--out',
        help='Path to the base output directory, e.g. out/debug-x86-64',
        required=True)
    parser.add_argument(
        '--tree',
        help='Restrict analysis to a source subtree, e.g. //apps/sysui/*',
        default='*')
    args, extras = parser.parse_known_args()

    # Ask gn about all the dart analyzer scripts.
    scripts = []
    targets = gn_describe(args.out, args.tree)
    for target_name, properties in targets.items():
        if ('type' not in properties or
                properties['type'] != 'action' or
                'script' not in properties or
                properties['script'] != '//build/dart/gen_analyzer_invocation.py' or
                'outputs' not in properties or
                not len(properties['outputs'])):
            continue
        script_path = properties['outputs'][0]
        script_path = script_path[2:]  # Remove the leading //
        scripts.append(os.path.join(paths.FUCHSIA_ROOT, script_path))

    # Put all the analyzer scripts in a queue that workers will work from
    script_queue = Queue.Queue()
    for script in scripts:
        script_queue.put(script)
    # Make a queue to receive results from workers.
    result_queue = Queue.Queue()
    # Track return codes from scripts.
    script_results = []
    failed_scripts = []

    # Create a worker thread for each CPU on the machine.
    for i in range(multiprocessing.cpu_count()):
        WorkerThread(script_queue, result_queue, extras).start()

    # Handle results from workers.
    while len(script_results) < len(scripts):
        script, returncode, output = result_queue.get(True)
        script_results.append(returncode)
        if returncode != 0:
            failed_scripts.append(script)
        print '----------------------------------------------------------'
        print output

    if len(failed_scripts):
        failed_scripts.sort()
        print 'Analysis failed in:'
        for script in failed_scripts:
            print '  %s' % script
        exit(1)


if __name__ == '__main__':
    sys.exit(main())
