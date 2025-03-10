# Minimum Permissions Finder (MPF) Design

The [High Level MPF flow](../Readme.MD#how-it-works) describes the overall flow of the Minimum Permissions Finder (MPF) system. This document provides a detailed design of the MPF system, including the key components and abstractions.

## Key Packages and Abstractions

![Key Interfaces and Packages](./images/mpf-key-interface.svg)

For each deployment type (ARM, Terraform) an implementation of the `DeploymentAuthorizationCheckerCleaner` interface is provided. The two key methods which need to be implemented for each deployment type implementation are GetDeploymentAuthorizationErrors() and CleanUpResources().

* The deployment type commands i.e. [armCmd](../cmd/armCmd.go), [bicepCmd](../cmd/bicepCmd.go), and [terraformCmd](../cmd/terraformCmd.go) are responsible for initializing the required dependencies including the `MPFService` to find the minimum permissions required for the deployment. This is illustrated in the sequence diagram below.
* [pkg/usecase/mpfService.go](../pkg/usecase/mpfService.go): Orchestrates the whole process of finding the minimum permissions required for any deployment type (ARM/Bicep/Terraform). It uses the `DeploymentAuthorizationCheckerCleaner` abstraction for any deployment type, be it ARM, Bicep, or Terraform. On receiving deployment authorization errors, it uses the `AuthorizationErrorParser` to parse the authorization errors and get the missing permissions and scopes. After adding the missing permissions to the custom role, it retries the deployment until it succeeds. It also cleans up all resources created during the process.
* [pkg/infrastructure/authorizationCheckers/ARMTemplateWhatIf/armTemplateWhatIfAuthorizationChecker.go](../pkg/infrastructure/authorizationCheckers/ARMTemplateWhatIf/armTemplateWhatIfAuthorizationChecker.go): Contains the DeploymentAuthorizationCheckerCleaner implementation for ARM (and Bicep) deployments.
* [pkg/infrastructure/authorizationCheckers/terraform/terraformAuthorizationChecker.go](../pkg/infrastructure/authorizationCheckers/terraform/terraformAuthorizationChecker.go): Contains the DeploymentAuthorizationCheckerCleaner implementation for Terraform deployments.
* [pkg/domain/authorizationErrorParser.go](../pkg/domain/authorizationErrorParser.go): Contains the core logic for the MPF, which is to parse the different kinds of authorization errors and figure out the required permissions and scopes from those errors.

## ARM and Terraform Sequence Diagrams

### ARM Sequence Diagram

For Bicep, the only difference with the ARM template flow is that as a first step the bicepCmd converts the Bicep file to an ARM template file.

![ARM Sequence Diagram](./images/arm-sequence.svg)

### Terraform Sequence Diagram

![Terraform Sequence Diagram](./images/terraform-sequence.svg)
