---
title: "Implement proper mocking for enforce command tests"
labels: ["enhancement", "testing"]
---
The enforce command tests currently use placeholder implementations because we can't easily mock the IAM client interface. We should implement proper mocking to have more comprehensive tests.
