# Tasks: <change-name>

## Milestone 1: <Subsystem name from design.md>
[One [TEST:<level>] task per guarantee — exactly one test entry each, at the
lowest level that proves it]
- [ ] 1.1 **[TEST:unit]** G1: <guarantee, verbatim from the spec> — `[TestName]`
      Spec: <frozen test definition — [input → expected] table rows, expected
      values hardcoded, or exact test code in a fenced block>
      RED: `<test-command-to-run-specific-test>` → FAIL (right reason: behavior missing)
      IMPL: <the minimal code to add>
      GREEN: `<test-command-to-run-specific-test>` → PASS
- [ ] 1.2 **[WIRING]** <pure plumbing — route / DI / config, wiring check only>.
      Verify: `<command-to-verify-the-wiring>`
- [ ] 1.3 **[VERIFY]** Milestone Integration Verification — <subsystem> integrates end-to-end.
      Verify: `<command exercising the whole milestone>`
- [ ] 1.4 **[COMMIT]** `<prefix>: <milestone-1 summary>`

## Milestone 2: <Next subsystem>
- [ ] 2.1 **[TEST:integration]** G2: <guarantee> — `[TestName]`.
      Spec: <frozen test definition>
      RED: `<command>` → FAIL
      IMPL: <minimal code>
      GREEN: `<command>` → PASS
- [ ] 2.N **[COMMIT]** `<prefix>: <milestone-2 summary>`

## Completion
- [ ] Run full project test suite: `<exact command>`
- [ ] Guarantee checklist: every spec guarantee maps to exactly one passing test entry
- [ ] Mutation spot-check: break 2–3 core guarantees, confirm a test fails, restore
- [ ] Update tasks.md and archive change
