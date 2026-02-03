# Plan: Add More Bicep Samples (Issue #87)

## Current State Analysis

### What currently exists:
1. **Bicep samples in repo:**
   - `aks-private-subnet.bicep` + params file (AKS cluster with VNet setup)
   - `subscription-scope-create-rg.bicep` + params file (subscription-level deployment)
   - `invalid-bicep.bicep` (for error testing)

2. **Documentation samples:**
   - **installation-and-quickstart.md**: Basic Bicep example with `--verbose` flag using the AKS sample
   - **display-options.MD**: Basic Bicep example, but ALL other output format examples use ARM templates (JSON output, detailed output, etc.)
   - **commandline-flags-and-env-variables.md**: Good documentation of `--initialPermissions` flag with examples using Terraform, but NO Bicep examples

### Key gaps identified:
- ❌ NO Bicep examples with `--jsonOutput` / `--outputFormat json` flag in documentation
- ❌ NO Bicep examples using `--initialPermissions` flag (neither comma-separated nor JSON file formats)
- ❌ NO Bicep `--showDetailedOutput` examples (only ARM examples exist)
- ❌ All advanced output format examples use ARM; Bicep is only shown for basic text output

---

## Proposed Solution

### Phase 1: Create Sample Files (if needed)
- **Review existing Bicep samples** - The `aks-private-subnet` sample appears sufficient for most demos
- **Consider creating a new simpler Bicep sample** (e.g., basic Storage Account) as an alternative for clarity, OR reuse existing ones
- **Create `bicep-backend-permissions.json`** sample file - a JSON file demonstrating the permissions structure needed for initialPermissions flag with Bicep

### Phase 2: Update Documentation

**A. Update `installation-and-quickstart.md`**
   - Add a new section: "Bicep with JSON Output" showing `--outputFormat json` example
   - Show command and example output (like what's shown in display-options.MD for ARM)
   - Include both Linux/macOS shell and Windows PowerShell variants

**B. Update `display-options.MD`**
   - Add "### Bicep JSON Output" section (parallel to existing ARM section)
   - Include full command example with environment variables
   - Include sample JSON output showing the permissions structure
   - Add "### Bicep Detailed Output" section showing `--showDetailedOutput` flag with a Bicep example
   - Ensure Bicep examples match the detail level of ARM examples

**C. Update `commandline-flags-and-env-variables.md`**
   - In the "Initial Permissions" section, add Bicep examples alongside existing Terraform examples
   - Show **comma-separated format** example with Bicep:
     ```bash
     azmpf bicep --initialPermissions "Microsoft.Storage/storageAccounts/read,Microsoft.Storage/storageAccounts/write" --bicepFilePath ./samples/bicep/aks-private-subnet.bicep ...
     ```
   - Show **JSON file format** example with Bicep using `@backend-permissions.json`:
     ```bash
     azmpf bicep --initialPermissions @bicep-backend-permissions.json --bicepFilePath ./samples/bicep/aks-private-subnet.bicep ...
     ```
   - Create and reference the `samples/bicep/bicep-backend-permissions.json` file

### Phase 3: Content Details for Each Update

**For JSON Output examples:**
- Command with all environment variables set
- Show the `--outputFormat json` flag (verify flag name - docs show both `--jsonOutput` and `--outputFormat json`)
- Include realistic JSON output with the permissions array and optional permissionsByResourceScope structure
- Add brief explanation of when/why to use JSON output (parsing, automation, CI/CD pipelines)

**For initialPermissions examples:**
- Explain use case: Bicep deployments with dependencies (e.g., storage accounts for module sources)
- Show comma-separated format (simpler, for single or few permissions)
- Show JSON file format (better for complex scenarios with many permissions)
- Explain the JSON structure (`RequiredPermissions` object with empty string key containing array)
- Include step-by-step: create the JSON file, then run the command with `@filename`

**For Detailed Output:**
- Show command with `--showDetailedOutput` flag
- Demonstrate the output structure with permission breakdowns by resource
- Explain when this is useful (understanding per-resource requirements, compliance auditing)

---

## Implementation Order

1. Create `samples/bicep/bicep-backend-permissions.json` 
2. Update `commandline-flags-and-env-variables.md` (add Bicep initialPermissions examples)
3. Update `display-options.MD` (add Bicep JSON output and detailed output sections)
4. Update `installation-and-quickstart.md` (add Bicep JSON output example)
5. Verify all examples are syntactically correct and use consistent sample files
6. Test commands locally if possible to ensure accuracy

---

## Deliverables Checklist

- [ ] New JSON permissions file for Bicep backend scenario
- [ ] Documentation sections for: Bicep JSON output, Bicep detailed output, Bicep initialPermissions (both formats)
- [ ] Examples for both Linux/macOS and Windows PowerShell
- [ ] Sample command outputs (can be real or representative)
- [ ] Clear explanations of when to use each feature
- [ ] Consistency with existing ARM template examples

---

## Notes

- This is a documentation-focused task. No changes to the MPF core code are required.
- The existing tool already supports all these features for Bicep; the documentation just needs to demonstrate them with concrete examples.
- Related closed PR #95 ("Add Bicep JSON output sample to documentation") may provide insights into what was already attempted.
