---
name: coops-tdd
description: Interactive behaviour-driven TDD with human approval gates. Use when writing tests, implementing features, fixing bugs, refactoring code, or making any behavioural change to the codebase with a human in the loop. Skip for: explaining code, reading files, documentation edits, config changes.
metadata:
  owner: will-head
  attribution:
    - "Ian Cooper: TDD, Where Did It All Go Wrong? (NDC Oslo, 2013); TDD Revisited (NDC Porto, 2023) — public-interface testing, behaviour as trigger, avoiding class-isolation mocks"
    - "Kent Beck: Test Driven Development: By Example (2002) — Red/Green/Refactor, Evident Data, test isolation definition"
    - "Dan North: Introducing BDD (2006) — behaviour vocabulary, when_[condition]_should_[outcome] naming"
    - "Bill Wake: Arrange, Act, Assert (2001) — AAA test structure"
    - "Gerard Meszaros: xUnit Test Patterns (2007) — Test Double taxonomy, Fake vs Mock distinction"
---

# Behaviour-Driven TDD — Interactive Mode

## Before You Start

Detect the project's test runner from config files (`package.json`, `pom.xml`, `Makefile`, etc.). If ambiguous, ask. Verify by running the suite once and confirming it exits cleanly.

If the `coding-standards` skill is available, invoke it to load standards for the language(s) being worked on. These will be applied during the Refactor phase. If unavailable, proceed without pre-loaded standards.

## Writing New Code: Red → Green → Refactor

1. **Red** — Write a failing test that describes the desired behaviour from the caller's perspective. Test at the public interface only. Run it. Confirm it fails for the right reason.
2. **Approval Gate** — STOP. Show the test to the user. Do NOT proceed without explicit user approval.
3. **Green** — Write the minimum code to make the test pass. Speed over design. No speculative code.
4. **Refactor** — Improve the implementation. Do NOT modify or add tests. Do NOT change behaviour. Run all tests after each change. Apply coding standards during this phase — standards compliance belongs here, not in Green. Green stays minimal. If `coding-standards` was loaded at session start, treat those rules as mandatory constraints. If it was unavailable, apply general code quality principles (clarity, consistency, no duplication).

## Scope Control

Each test should be the most obvious, smallest step toward the requirement. If you find yourself writing a lot of code to make one test pass, the test is probably too large — break it into a smaller first step. Only add code needed to satisfy a behavioural requirement expressed in a test.

## Modifying Existing Code

1. Run the full test suite. Confirm all tests pass.
2. Make the change.
3. Run the full test suite again. All tests must still pass.
4. If tests fail, the implementation is wrong — revert and try again.

## Test Rules

- Never write production code except to make a failing test pass.
- Only write a test in response to a new behaviour or requirement — never in response to a new method or class.
- Tests must come from user requirements — do not invent scenarios.
- Test at the public interface only. Never test private or internal methods.
- Never expose internals just to test them.
- **Never modify existing tests to make implementation changes pass. This is reward hacking.**
- Tests must be fast (seconds, not minutes) and binary (pass/fail, no interpretation needed).

### Test Naming

`when_[condition]_should_[outcome]` — adapt to language conventions (snake_case, PascalCase, etc.):
- `when_balance_is_zero_should_reject_withdrawal`
- `when_email_is_invalid_should_raise_error`

### Test Structure — Arrange / Act / Assert

```python
def test_when_balance_is_zero_should_reject_withdrawal():
    # Arrange
    account = Account(balance=0)

    # Act / Assert
    with pytest.raises(InsufficientFundsError):
        account.withdraw(10)
```

Use Evident Data: only include values that affect the test outcome. Use builders or helpers to hide irrelevant setup.

## Mock Rules

- Do NOT use mocks by default.
- Only mock slow I/O (network, database, filesystem, message queues).
- Never mock internal collaborators to isolate a class.
- Prefer in-memory implementations over mocks.

## Refactoring Rules

- Refactoring = changing implementation without changing behaviour.
- During refactoring, existing tests MUST NOT be modified or deleted.
- New classes or methods extracted during refactoring do not get their own tests — they are covered via the public interface.
- If tests break during refactoring, flag this to the user — they were coupled to implementation details.

## What Not To Do

If you catch yourself doing any of these, stop and revert:

- Writing tests after implementation rather than before.
- Modifying or deleting existing tests to make implementation changes pass.
- Writing speculative code not required by any test.
- Writing a test in response to a new method or class rather than a new behaviour.

For reasoning behind these rules, see `references/tdd-philosophy.md`.
