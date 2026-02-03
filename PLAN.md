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

### Phase 1: Create Sample Files

#### 1.1: Create a Simple Storage Account Bicep Sample
We will create a **new, simpler Bicep example** focused on a basic Azure Storage Account deployment. This provides:
- **Simplicity**: Much easier to understand than AKS (which has VNets, subnets, clusters)
- **Real-world relevance**: Storage accounts are commonly deployed and a natural use case for backend state
- **Clear permissions output**: Storage account deployments generate a focused set of permissions that are easy to demonstrate

**Files to create:**
- `samples/bicep/storage-account-simple.bicep` - Basic storage account with minimal parameters
- `samples/bicep/storage-account-simple-params.json` - Parameters file for the storage account

**Purpose in documentation:**
- Used for basic output format examples (text, JSON, detailed output)
- Fast execution (~1-2 minutes) makes it ideal for documentation examples
- Clear permission output (4-5 permissions) easier to read than complex templates
- Can reuse across multiple documentation sections

#### 1.2: Use Existing AKS Bicep Sample for `--initialPermissions` Demonstration
We will use the **existing `samples/bicep/aks-private-subnet.bicep`** template to demonstrate the `--initialPermissions` flag.

**Why AKS for this use case:**
- Represents a realistic scenario: AKS cluster depends on backend storage account
- Shows the real value of `--initialPermissions`: reduces execution time (5-10 min → 2-3 min)
- Complex enough to meaningfully benefit from pre-seeded permissions
- Already exists in repo (no new code to create)

**Scenario explanation:**
The AKS cluster is deploying, but assumes a pre-existing backend storage account is available (e.g., for state, configuration, or secrets). By pre-seeding storage permissions with `--initialPermissions`, we avoid MPF wasting iterations on storage permission discovery and focus only on AKS requirements.

#### 1.3: Create Backend Permissions JSON File
Create `samples/bicep/bicep-backend-permissions.json` - a JSON file demonstrating the permissions structure needed for the `--initialPermissions` flag with Bicep deployments that use remote backends.

**File contents:**
```json
{
  "RequiredPermissions": {
    "": [
      "Microsoft.Storage/storageAccounts/read",
      "Microsoft.Storage/storageAccounts/listKeys/action",
      "Microsoft.Storage/storageAccounts/blobServices/containers/read",
      "Microsoft.Storage/storageAccounts/blobServices/containers/blobs/read",
      "Microsoft.Storage/storageAccounts/blobServices/containers/blobs/write"
    ]
  }
}
```

**Why this file:**
- Provides a concrete example referenced in documentation
- Shows real-world scenario: Bicep deployment that needs to access state in Azure Storage
- Demonstrates both what the JSON structure looks like and when it's needed
- Can be referenced when showing both storage account and AKS examples

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

## Phase 1 Implementation Steps

1. ✅ Create `samples/bicep/storage-account-simple.bicep` with the storage account Bicep code
2. ✅ Create `samples/bicep/storage-account-simple-params.json` with parameters
3. ✅ Add comprehensive comments to storage-account-simple.bicep for junior developers
4. ✅ Create `samples/bicep/bicep-backend-permissions.json` with the permissions structure
5. ✅ Commit these files to the branch with message: "docs(samples): add storage account bicep example for issue #87"
6. ✅ Identify existing `samples/bicep/aks-private-subnet.bicep` for use in `--initialPermissions` documentation examples
7. ✅ Plan: Use `bicep-backend-permissions.json` to show pre-seeding storage permissions for AKS deployment scenario

## Phase 2 & 3 Implementation Order (Future)

2. Update `commandline-flags-and-env-variables.md` (add Bicep initialPermissions examples)
3. Update `display-options.MD` (add Bicep JSON output and detailed output sections)
4. Update `installation-and-quickstart.md` (add Bicep JSON output example)
5. Verify all examples are syntactically correct and use consistent sample files
6. Test commands locally if possible to ensure accuracy

---

## Deliverables Checklist

## Phase 1 Documentation Samples Summary

| Sample | Purpose | Execution Time | Output Complexity | Use Cases |
|--------|---------|-----------------|-------------------|-----------|
| **storage-account-simple.bicep** | Basic output format demos | ~1-2 min | 4-5 permissions | Text output, JSON output, Detailed output examples |
| **aks-private-subnet.bicep** (existing) | `--initialPermissions` demo | ~5-10 min | 9+ permissions | Showing time savings with pre-seeded permissions |
| **bicep-backend-permissions.json** | Pre-seeded permissions example | N/A | Reference file | Demonstrating permission structure and use case |

### Phase 2 & 3 (Future)
- [ ] New JSON permissions file for Bicep backend scenario *(Phase 1)*
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
