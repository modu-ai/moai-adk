#!/bin/sh
# AC-CL-008 combine-rule fixture: both streams carry a non-empty line.
printf 'noise from stdout\n'
printf 'Logged in using ChatGPT\n' >&2
