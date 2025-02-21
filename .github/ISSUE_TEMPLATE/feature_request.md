# yaml-language-server: $schema=https://json.schemastore.org/github-issue-forms.json
# https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/syntax-for-githubs-form-schema
---
name: Feature Request
description: Make a feature request
labels: ["enhancement"]
body:
  - id: description
    type: textarea
    attributes:
      label: Description
      placeholder: Please provide a succinct description of the feature you would like to see.
    validations:
      required: true