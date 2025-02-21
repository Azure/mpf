# yaml-language-server: $schema=https://json.schemastore.org/github-issue-forms.json
# https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/syntax-for-githubs-form-schema
---
name: Bug Report
description: File a bug report.
title: "[Bug]: "
labels: ["bug"]
body:
  - type: markdown
    attributes:
      value: |
        Thanks for taking the time to fill out this bug report!
  - type: dropdown
    id: version
    attributes:
      label: Version
      description: What version of MPF/bicep/terraform are you running. Please include OS/distribution?
    validations:
      required: true
  - type: textarea
    id: Description
    attributes:
      label: What happened?
      description: Also tell us, what did you expect to happen?
      placeholder: Tell us what you see!
      value: "A bug happened!"
    validations:
      required: true
  - type: textarea
    id: Steps
    attributes:
      label: Steps to reproduce
      description: Please provide detailed steps to reproduce the bug.
      placeholder: Tell us what you did!
      value: "I did this, then that, then this happened!"
    validations:
      required: true
   - type: textarea
    id: logs
    attributes:
      label: Relevant log output
      description: Please copy and paste any relevant log output. This will be automatically formatted into code, so no need for backticks.
      render: shell