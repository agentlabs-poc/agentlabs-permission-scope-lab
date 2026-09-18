#!/usr/bin/env python3
"""Refuse a rule identifier claimed by more than one question.

Every decision in docs/ is written as "Q-nnn / FAMILY-nnn". The families are
registries, and nothing enforced that a number was free before it was used. In
one sitting four identifiers were assigned twice — DECISION-011 and -012 to
Q-063/Q-064 and again to Q-134/Q-135, PERMISSION-004 to Q-058, ENFORCEMENT-009
to Q-075, RESOLUTION-003 to role expansion. Each was caught by hand, which is
not a control.

Run from the repository root. Exits non-zero on a collision.

Scope: docs/*.md only. docs/history/ deliberately preserves superseded wording,
so an identifier legitimately appears there under its original question.
"""
import re, glob, sys
FAMILY = r'(?!Q-)([A-Z]{4,}-\d+)'
owners = {}
for p in glob.glob('docs/*.md'):          # current chapters only; history/ keeps superseded text
    s = open(p).read()
    for m in re.finditer(r'\bQ-(\d+) / ' + FAMILY, s):
        owners.setdefault(m.group(2), set()).add('Q-' + m.group(1))
bad = {k: sorted(v) for k, v in owners.items() if len(v) > 1}
for k, v in sorted(bad.items()):
    print("COLLISION", k, "claimed by", v)
print(f"\n{len(bad)} colliding rule identifiers" if bad else f"all {len(owners)} paired rule identifiers have a single owner")
sys.exit(1 if bad else 0)
