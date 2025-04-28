
variable "tags" {
  type        = map(string)
  default     = null
  description = "A map of tags to add to all resources"
}

variable "log_analytics_workspace_name" {
  type        = string
  default     = ""
  description = "The name of the Log Analytics Workspace. If not provided, a name will be generated."
}

