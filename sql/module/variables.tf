variable "region" {
  type        = string
  description = "Which region should the resources be provisioned in"
}

variable "database_name" {
  type        = string
  description = "What should we call the database we create?"
  default     = "apprepo"
}

variable "workgroup_name" {
  type        = string
  description = "What should we call the workgroup we create?"
  default     = "design-as-code"
}