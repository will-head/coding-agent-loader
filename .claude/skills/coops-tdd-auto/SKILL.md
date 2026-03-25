---
name: coops-tdd-auto
description: Behaviour-driven TDD for automated agents working from a task list. Use when implementing features, writing new code, fixing bugs, adding functions, building integrations, or making any behavioural change to the codebase, regardless of scale. Skip for: explaining code, reading files, documentation edits, config changes.
metadata:
  owner: will-head
  attribution:
    - "Ian Cooper: TDD, Where Did It All Go Wrong? (NDC Oslo, 2013); TDD Revisited (NDC Porto, 2023) — public-interface testing, behaviour as trigger, avoiding class-isolation mocks"
    - "Kent Beck: Test Driven Development: By Example (2002) — Red/Green/Refactor, Evident Data, test isolation definition"
    - "Dan North: Introducing BDD (2006) — behaviour vocabulary, when_[condition]_should_[outcome] naming"
    - "Bill Wake: Arrange, Act, Assert (2001) — AAA test structure"
    - "Gerard Meszaros: xUnit Test Patterns (2007) — Test Double taxonomy, Fake vs Mock distinction"
---

# Behaviour-Driven TDD — Automated Mode

TDD is mandatory. Do not write implementation code before writing a failing test.

## Before You Start

Detect the project's test runner from config files (`package.json`, `pom.xml`, `Makefile`, etc.). If ambiguous, infer from the language and project structure. Verify by running the suite once and confirming it exits cleanly.

If the `coding-standards` skill is available, invoke it to load standards for the language(s) being worked on. These will be applied during the Refactor phase. If unavailable, proceed without pre-loaded standards.

## Red → Green → Refactor

### Red — Write a Failing Test

1. Derive the behaviour from the current task item. One task item = one or more behaviours = one or more tests.
2. Write a test that specifies that behaviour from the caller's perspective.
3. Test at the public interface (exports, public methods, observable outcomes). Never test internals.
4. Run the test. Confirm it fails for the right reason — the behaviour is absent, not a syntax error or import problem.

If a task item maps to multiple distinct behaviours, write one test per behaviour — do not combine. If a task item is too vague to derive a testable behaviour, flag it rather than guessing.

**Test file naming** — one test file per behaviour where practical, named for the behaviour being tested.

**Naming** — `when_[condition]_should_[outcome]`, adapted to language conventions:
- `when_balance_is_zero_should_reject_withdrawal`
- `when_email_is_invalid_should_raise_error`
- `when_password_is_too_short_should_fail_validation`

**Structure — Arrange / Act / Assert:**

```python
def test_when_balance_is_zero_should_reject_withdrawal():
    # Arrange
    account = Account(balance=0)

    # Act / Assert
    with pytest.raises(InsufficientFundsError):
        account.withdraw(10)
```

Use Evident Data: only include values that affect the test outcome. Use builders or helpers to hide irrelevant setup noise.

### Green — Make the Test Pass

Write the **minimum code** to make the test pass. Nothing more. Speed over design — cleanup is for Refactor.

Do not write code for requirements not expressed in a test.

### Refactor — Improve the Design

With tests green, improve structure without changing behaviour:
- Rename, extract, reorganise — do not change what the code does.
- Run all tests after each change.
- Do NOT modify or add tests during refactoring.
- Apply coding standards during this phase — standards compliance belongs here, not in Green. Green stays minimal. If `coding-standards` was loaded at session start, treat those rules as mandatory constraints. If it was unavailable, apply general code quality principles (clarity, consistency, no duplication).

Repeat the cycle for the next behaviour.

## Scope Control

Each test should be the most obvious, smallest step toward the requirement. If you find yourself writing a lot of code to make one test pass, the test is probably too large — break it into a smaller first step. Only add code needed to satisfy a behavioural requirement expressed in a test.

## Modifying Existing Code

1. Run the full test suite. Confirm all tests pass.
2. Make the change.
3. Run the full test suite again. All tests must still pass.
4. If tests fail, the implementation is wrong — revert and try again. Do not modify tests to compensate.

## Test Rules

- Never write production code except to make a failing test pass.
- Tests must come from task requirements. Do not invent scenarios not specified by the task.
- Only write a test in response to a new behaviour — never in response to a new method or class.
- Test at the public interface only. Never test private or internal methods or classes.
- Never expose internals just to test them.
- **Never modify existing tests to make implementation changes pass. This is reward hacking.**
- Tests must be fast (seconds, not minutes) and binary (pass/fail, no interpretation needed).
- Code coverage is a tool for guiding refactoring, not a target.

## Mock Rules

- Do NOT mock internal collaborators to isolate classes.
- Only use test doubles for slow I/O (network, database, filesystem, message queues).
- Prefer in-memory implementations over mocks — they are more honest about behaviour.

## Refactoring Rules

- Refactoring = changing implementation without changing behaviour.
- During refactoring, existing tests MUST NOT be modified or deleted.
- New classes or methods extracted during refactoring do not get their own tests — they are covered via the public interface.
- Refactoring must never break tests — well-written tests verify behaviour, not implementation. If tests break during refactoring, they were coupled to implementation details. Revert the refactor, log the coupling issue in the task output, and continue to the next task item. Do not modify the tests.

## What Not To Do

If you catch yourself doing any of these, stop and revert:

- Writing tests after implementation rather than before.
- Modifying or deleting existing tests to make implementation changes pass.
- Writing speculative code not required by any test.
- Writing a test in response to a new method or class rather than a new behaviour.

For reasoning behind these rules, see `references/tdd-philosophy.md`.
