# Collaboration Workflow

This document defines the working flow for planning, implementing, and closing feature or fix work in the courier project.

## Goals

- Keep requirements clear before implementation starts.
- Capture agreed work in GitHub issues so the discussion has a durable source of truth.
- Implement in small, reviewable steps.
- Update and close the related issue only after the user confirms the result meets the goal.

## Standard Flow

1. Problem or feature intake

The user describes a problem, requirement, bug, or feature idea. The assistant should first understand the context, inspect relevant code when needed, and avoid jumping directly into edits if the user is still exploring options.

2. Discussion and approach selection

The assistant and user discuss the solution, tradeoffs, implementation boundaries, affected services, data model impact, gRPC/API contract impact, migration needs, and validation steps. The output of this stage should be a clear agreed plan.

3. GitHub issue creation or update

After the approach is agreed, the assistant should use GitHub access to create or update an issue. The issue should include:

- Problem statement
- Context and motivation
- Agreed solution
- Implementation steps
- Affected services or modules
- Migration or configuration impact, if any
- Validation checklist
- Completion criteria

4. Implementation

The assistant implements the agreed work in focused steps. Before broad edits, the assistant should explain what files or modules will be touched. For generated files such as Wire, Swagger, protobuf, or vendor outputs, regenerate using the project commands instead of editing generated code manually.

5. Verification

The assistant verifies the change with the relevant project commands. If a command cannot be run because of missing tools, environment limits, or external services, the assistant should state that clearly and provide the exact command for the user to run.

6. User confirmation

The user reviews and confirms whether the implementation meets the requirement. The assistant should not close the related issue before this confirmation unless the user explicitly asks to do so.

7. Issue update and closure

After confirmation, the assistant comments on the GitHub issue with:

- Final solution summary
- Important implementation details
- Validation performed
- Any follow-up work intentionally left for another ticket

Then the assistant closes the issue as completed.

## Issue Comment Template

```md
Update status: resolved.

Implemented solution:
- ...
- ...
- ...

Validation performed:
- ...
- ...

Notes:
- ...

Closing this issue as completed.
```

## Operating Rules

- Do not create an issue until the approach is agreed.
- Do not make code changes during a planning-only discussion.
- Do not close an issue until the user confirms the objective is complete.
- Prefer one issue per coherent feature or bug.
- Create follow-up issues for intentionally deferred work instead of expanding the current scope silently.
- Keep GitHub comments concise but specific enough for future readers to understand what changed.
