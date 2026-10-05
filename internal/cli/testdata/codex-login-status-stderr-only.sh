#!/bin/sh
# AC-CL-008 combine-rule fixture: the status line goes ENTIRELY to stderr.
# This is the shape of the production defect this SPEC repairs (§A.2).
printf 'Logged in using ChatGPT\n' >&2
