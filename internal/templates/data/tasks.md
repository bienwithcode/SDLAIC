# Tasks: <change-name>

## Milestone 1: <Subsystem name from design.md>
[TDD task pairs for this subsystem, in order]
- [ ] 1.1 **[TEST-RED:unit]** Add `[TestName]` — assert [behavior].
      `<test-command-to-run-specific-test>` → FAIL
- [ ] 1.2 **[IMPL]** Add `[ImplementationName]` to pass 1.1.
      GREEN: `<test-command-to-run-specific-test>` → PASS
- [ ] 1.3 **[VERIFY]** Milestone Integration Verification — <subsystem> integrates end-to-end.
      Verify: `<command exercising the whole milestone>`
- [ ] 1.4 **[COMMIT]** `<prefix>: <milestone-1 summary>`

## Milestone 2: <Next subsystem>
- [ ] 2.1 **[TEST-RED:unit]** Add `[TestName]` — assert [behavior].
      `<test-command-to-run-specific-test>` → FAIL
- [ ] 2.2 **[IMPL]** Add `[ImplementationName]` to pass 2.1.
      GREEN: `<test-command-to-run-specific-test>` → PASS
- [ ] 2.N **[COMMIT]** `<prefix>: <milestone-2 summary>`

## Completion
- [ ] Run full project test suite: `<exact command>`
- [ ] Manual QA against Jira acceptance criteria
- [ ] Update tasks.md and archive change

## Challenge & Resolution Log
<!-- From the plan grill. State "No grill (workflow: <level>)" if none ran. -->
| Challenge | Resolution |
|-----------|------------|
| [sizing / TDD / milestone concern] | [how it was resolved] |
