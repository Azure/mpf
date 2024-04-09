# Debugging Locally

The [launch.json](.vscode/launch.json) file has been configured to debug the application locally. The values for the required global arguments are passed via the .env.sh file. The following is a sample .env.sh file:

```shell
MPF_SUBSCRIPTIONID=YOUR_SUBSCRIPTION_ID
MPF_TENANTID=YOUR_TENANT_ID
MPF_SPCLIENTID=YOUR_SP_CLIENT_ID
MPF_SPCLIENTSECRET=YOUR_SP_CLIENT_SECRET
MPF_SPOBJECTID=YOUR_SP_OBJECT_ID
```

Next the arguments can me modified in the launch.json file for ARM or Terraform as shown:

* ARM

```json
"args": ["arm", "--templateFilePath", "PATH_TO_ARM_TEMPLATE_FILE", "--parametersFilePath", "PATH_TO_PARAMETERS_FILE"]
```

* Terraform

```json
"args": ["terraform", "--tfPath", "PATH_TO_TERRAFORM_EXECUTABLE" , "--workingDir", "PATH_TO_TERRAFORM_MODULE_DIRECTORY", "--varFilePath", "PATH_TO_TERRAFORM_VARIABLES_FILE"]
```

These values need to be modified to match your environment. You can also modify the launch.json to change the logging verbosity, which is debug by default while debugging.

