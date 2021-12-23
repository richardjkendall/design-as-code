provider "aws" {
  region = var.region
}

data "aws_caller_identity" "current" {}

data "aws_iam_policy_document" "query_user_policy" {
  statement {
    sid     = "glue"
    effect  = "Allow"
    actions = [
      "glue:GetTable",
      "glue:GetTables",
      "glue:GetDatabases"
    ]
    resources = [
      "arn:aws:glue:${var.region}:${data.aws_caller_identity.current.account_id}:catalog",
      "arn:aws:glue:${var.region}:${data.aws_caller_identity.current.account_id}:database/${aws_athena_database.apprepo.id}",
      "arn:aws:glue:${var.region}:${data.aws_caller_identity.current.account_id}:table/${aws_athena_database.apprepo.id}/*"
    ]
  }
  
  statement {
    sid     = "athenaall"
    effect  = "Allow"
    actions = [
      "athena:ListEngineVersions",
      "athena:ListWorkGroups",
      "athena:ListDataCatalogs",
      "athena:ListDatabases",
      "athena:GetDatabase",
      "athena:ListTableMetadata",
      "athena:GetTableMetadata"
    ]
    resources = [
      "*"
    ]
  }

  statement {
    sid     = "athenaworkgroup"
    effect  = "Allow"
    actions = [
      "athena:GetWorkGroup", 
      "athena:BatchGetQueryExecution",
      "athena:GetQueryExecution",
      "athena:ListQueryExecutions",
      "athena:StartQueryExecution",
      "athena:StopQueryExecution",
      "athena:GetQueryResults",
      "athena:GetQueryResultsStream",
      "athena:CreateNamedQuery",
      "athena:GetNamedQuery",
      "athena:BatchGetNamedQuery",
      "athena:ListNamedQueries",
      "athena:DeleteNamedQuery",
      "athena:CreatePreparedStatement",
      "athena:GetPreparedStatement",
      "athena:ListPreparedStatements",
      "athena:UpdatePreparedStatement",
      "athena:DeletePreparedStatement"
    ]
    resources = [
      aws_athena_workgroup.design_as_code_workgroup.arn
    ]
  }

  statement {
    sid     = "s3sourcelist"
    effect  = "Allow"
    actions = [
      "s3:ListBucket"
    ]
    resources = [
      aws_s3_bucket.source_bucket.arn
    ]
  }

  statement {
    sid     = "s3sourceget"
    effect  = "Allow"
    actions = [
      "s3:GetObject"
    ]
    resources = [
      "${aws_s3_bucket.source_bucket.arn}/*"
    ]
  }

  statement {
    sid     = "s3results"
    effect  = "Allow"
    actions = [
      "s3:GetBucketLocation",
      "s3:GetObject",
      "s3:ListBucket",
      "s3:ListBucketMultipartUploads",
      "s3:AbortMultipartUpload",
      "s3:PutObject",
      "s3:ListMultipartUploadParts"
    ]
    resources = [
      "${aws_s3_bucket.results_bucket.arn}/*",
      "${aws_s3_bucket.results_bucket.arn}"
    ]
  }

  statement {
    sid     = "kms"
    effect  = "Allow"
    actions = [
      "kms:Decrypt",
      "kms:GenerateDataKey"
    ]
    resources = [
      aws_kms_key.source_bucket_key.arn
    ]
  }

}

resource "aws_iam_policy" "query_user_policy" {
  name_prefix = "design-as-code-query"
  policy      = data.aws_iam_policy_document.query_user_policy.json
}

resource "aws_s3_bucket" "results_bucket" {
  bucket_prefix = "athena-results"
  force_destroy = true
}

resource "aws_s3_bucket_public_access_block" "block_results_pub_access" {
  bucket = aws_s3_bucket.results_bucket.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_athena_database" "apprepo" {
  name   = "appreponp"
  bucket = aws_s3_bucket.results_bucket.bucket

  encryption_configuration {
    encryption_option = "SSE_KMS"
    kms_key           = aws_kms_key.source_bucket_key.arn
  } 
}

resource "aws_athena_workgroup" "design_as_code_workgroup" {
  name          = "design-as-code"
  force_destroy = true

  configuration {
    enforce_workgroup_configuration    = true
    publish_cloudwatch_metrics_enabled = true

    result_configuration {
      output_location = "s3://${aws_s3_bucket.results_bucket.bucket}"

      encryption_configuration {
        encryption_option = "SSE_KMS"
        kms_key_arn       = aws_kms_key.source_bucket_key.arn
      }
    }
  }
}

resource "aws_kms_key" "source_bucket_key" {
  description             = "KMS key for design-as-code buckets"
  deletion_window_in_days = 30
}

resource "aws_s3_bucket" "source_bucket" {
  bucket_prefix = "design-as-code"

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        kms_master_key_id = aws_kms_key.source_bucket_key.arn
        sse_algorithm     = "aws:kms"
      }
    }
  }

}

resource "aws_s3_bucket_public_access_block" "block_source_pub_access" {
  bucket = aws_s3_bucket.source_bucket.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_glue_catalog_table" "ummatched_resources_view" {
  name          = "unmatched_resources"
  database_name = aws_athena_database.apprepo.id

  table_type = "VIRTUAL_VIEW"
  
  parameters = {
    presto_view = "true"
  }

  storage_descriptor {
   
    ser_de_info {
      name = "blank"
    } 

    columns {
      name = "solutionname"
      type = "string"
    }
    
    columns {
      name = "solutionnumber"
      type = "string"
    }

    columns {
      name = "resourcename"
      type = "string"
    }

    columns {
      name = "resourcetype"
      type = "string"
    }

  }

  view_original_text = join(" ", ["/* Presto View:", base64encode(templatefile("${path.module}/unmatched_resources.json", {
    database_name = aws_athena_database.apprepo.id
  })), "*/"])
  view_expanded_text = "/* Presto View */"

  depends_on = [
    aws_glue_catalog_table.only_newest_matches_view
  ]
}

resource "aws_glue_catalog_table" "solution_summary_view" {
  name          = "solution_summary"
  database_name = aws_athena_database.apprepo.id

  table_type = "VIRTUAL_VIEW"
  
  parameters = {
    presto_view = "true"
  }

  storage_descriptor {
   
    ser_de_info {
      name = "blank"
    } 
   

    columns {
      name = "solutionname"
      type = "string"
    }
    
    columns {
      name = "solutionnumber"
      type = "string"
    }

    columns {
      name = "solution_count"
      type = "int"
    }

    columns {
      name = "total_resources"
      type = "bigint"
    }

    columns {
      name = "total_matching"
      type = "bigint"
    }

    columns {
      name = "total_not_matching"
      type = "bigint"
    }   

  }

  view_original_text = join(" ", ["/* Presto View:", base64encode(templatefile("${path.module}/solution_summary.json", {
    database_name = aws_athena_database.apprepo.id
  })), "*/"])
  view_expanded_text = "/* Presto View */"

  depends_on = [
    aws_glue_catalog_table.only_newest_matches_view
  ]
}

resource "aws_glue_catalog_table" "pattern_summary_view" {
  name          = "pattern_summary"
  database_name = aws_athena_database.apprepo.id

  table_type = "VIRTUAL_VIEW"
  
  parameters = {
    presto_view = "true"
  }

  storage_descriptor {
   
    ser_de_info {
      name = "blank"
    } 
   
    columns {
      name = "patternname"
      type = "string"
    }
    
    columns {
      name = "resourcetype"
      type = "string"
    }
    
    columns {
      name = "solutionnumber"
      type = "string"
    }

    columns {
      name = "total_resources"
      type = "bigint"
    }

  }

  view_original_text = join(" ", ["/* Presto View:", base64encode(templatefile("${path.module}/pattern_summary.json", {
    database_name = aws_athena_database.apprepo.id
  })), "*/"])
  view_expanded_text = "/* Presto View */"

  depends_on = [
    aws_glue_catalog_table.only_newest_matches_view
  ]
}

resource "aws_glue_catalog_table" "resources_with_attr_view" {
  name          = "resources_with_attr"
  database_name = aws_athena_database.apprepo.id

  table_type = "VIRTUAL_VIEW"
  
  parameters = {
    presto_view = "true"
  }

  storage_descriptor {
   
    ser_de_info {
      name = "blank"
    } 
   
    columns {
      name = "version"
      type = "string"
    }

    columns {
      name = "solutionname"
      type = "string"
    }

    columns {
      name = "solutionnumber"
      type = "string"
    }

    columns {
      name = "resourcetype"
      type = "string"
    }

    columns {
      name = "resourcename"
      type = "string"
    }

    columns {
      name = "dependson"
      type = "array<string>"
    }
    
    columns {
      name = "name"
      type = "string"
    }
    
    columns {
      name = "value"
      type = "string"
    }
    
  }

  view_original_text = join(" ", ["/* Presto View:", base64encode(templatefile("${path.module}/resources_with_attr.json", {
    database_name = aws_athena_database.apprepo.id
  })), "*/"])
  view_expanded_text = "/* Presto View */"

  depends_on = [
    aws_glue_catalog_table.only_newest_resources_view
  ]
}

resource "aws_glue_catalog_table" "only_newest_resources_view" {
  name          = "only_newest_resources"
  database_name = aws_athena_database.apprepo.id

  table_type = "VIRTUAL_VIEW"
  
  parameters = {
    presto_view = "true"
  }

  storage_descriptor {
   
    ser_de_info {
      name = "blank"
    } 
   
    
    columns {
      name = "version"
      type = "string"
    }
    
    columns {
      name = "solutionname"
      type = "string"
    }
    
    columns {
      name = "solutionnumber"
      type = "string"
    }

    columns {
      name = "resourcename"
      type = "string"
    }
    
    columns {
      name = "resourcetype"
      type = "string"
    }
    
    columns {
      name = "dependson"
      type = "array<string>"
    }
    
    columns {
      name = "attributes"
      type = "array<struct<name:string,value:string>>"
    } 
    
    columns {
      name = "solnumber"
      type = "string"
    } 
    
    columns {
      name = "currentver"
      type = "string"
    }

  }

  view_original_text = join(" ", ["/* Presto View:", base64encode(templatefile("${path.module}/only_newest_resources.json", {
    database_name = aws_athena_database.apprepo.id
  })), "*/"])
  view_expanded_text = "/* Presto View */"

  depends_on = [
    aws_glue_catalog_table.newest_resources_view,
    aws_glue_catalog_table.resources_table
  ]

}

resource "aws_glue_catalog_table" "only_newest_matches_view" {
  name          = "only_newest_matches"
  database_name = aws_athena_database.apprepo.id

  table_type = "VIRTUAL_VIEW"
  
  parameters = {
    presto_view = "true"
  }

  storage_descriptor {
   
    ser_de_info {
      name = "blank"
    } 
   
    
    columns {
      name = "version"
      type = "string"
    }
    
    columns {
      name = "solutionname"
      type = "string"
    }
    
    columns {
      name = "solutionnumber"
      type = "string"
    }

    columns {
      name = "resourcename"
      type = "string"
    }
    
    columns {
      name = "resourcetype"
      type = "string"
    }
    
    columns {
      name = "dependson"
      type = "array<string>"
    }
  
    columns {
      name = "patternname"
      type = "string"
    }
    
    columns {
      name = "patterntarget"
      type = "string"
    }
                
    columns {
      name = "matchespattern"
      type = "boolean"
    }
    
    columns {
      name = "attributes"
      type = "array<struct<name:string,value:string>>"
    } 
    
    columns {
      name = "solnumber"
      type = "string"
    } 
    
    columns {
      name = "currentver"
      type = "string"
    }

  }

  view_original_text = join(" ", ["/* Presto View:", base64encode(templatefile("${path.module}/only_newest_matches.json", {
    database_name = aws_athena_database.apprepo.id
  })), "*/"])
  view_expanded_text = "/* Presto View */"

  depends_on = [
    aws_glue_catalog_table.newest_matches_view,
    aws_glue_catalog_table.matches_table
  ]

}

resource "aws_glue_catalog_table" "newest_resources_view" {
  name          = "newest_resources"
  database_name = aws_athena_database.apprepo.id

  table_type = "VIRTUAL_VIEW"
  
  parameters = {
    presto_view = "true"
  }

  storage_descriptor {
   
    ser_de_info {
      name = "blank"
    } 
   
    columns {
      name = "solnumber"
      type = "string"
    }

    columns {
      name = "currentver"
      type = "string"
    }
  }

  view_original_text = join(" ", ["/* Presto View:", base64encode(templatefile("${path.module}/newest_resources.json", {
    database_name = aws_athena_database.apprepo.id
  })), "*/"])
  view_expanded_text = "/* Presto View */"

  depends_on = [
    aws_glue_catalog_table.resources_table
  ]

}

resource "aws_glue_catalog_table" "newest_matches_view" {
  name          = "newest_matches"
  database_name = aws_athena_database.apprepo.id

  table_type = "VIRTUAL_VIEW"
  
  parameters = {
    presto_view = "true"
  }

  storage_descriptor {
   
    ser_de_info {
      name = "blank"
    } 
   
    columns {
      name = "solnumber"
      type = "string"
    }

    columns {
      name = "currentver"
      type = "string"
    }
  }

  view_original_text = join(" ", ["/* Presto View:", base64encode(templatefile("${path.module}/newest_matches.json", {
    database_name = aws_athena_database.apprepo.id
  })), "*/"])
  view_expanded_text = "/* Presto View */"

  depends_on = [
    aws_glue_catalog_table.matches_table
  ]

}

resource "aws_glue_catalog_table" "resources_table" {
  name          = "resources"
  database_name = aws_athena_database.apprepo.id

  table_type = "EXTERNAL_TABLE"

  storage_descriptor {
    location      = "s3://${aws_s3_bucket.source_bucket.id}/resources"
    input_format  = "org.apache.hadoop.mapred.TextInputFormat"
    output_format = "org.apache.hadoop.hive.ql.io.IgnoreKeyTextOutputFormat"
    
    ser_de_info {
      name = "test"
      serialization_library = "org.openx.data.jsonserde.JsonSerDe"
      
      parameters = {
        "serialization.format" = 1
      }
      
    }

    columns {
       name = "version"
       type = "string"
    }

    columns {
       name = "solutionname"
       type = "string"
    }

    columns {
       name = "solutionnumber"
       type = "string"
    }

    columns {
       name = "resourcename"
       type = "string"
    }

    columns {
       name = "resourcetype"
       type = "string"
    }

    columns {
       name = "dependson"
       type = "array<string>"
    }

    columns {
       name = "attributes"
       type = "array<struct<name:string,value:string>>"
    }

  }

}

resource "aws_glue_catalog_table" "matches_table" {
  name          = "matches"
  database_name = aws_athena_database.apprepo.id

  table_type = "EXTERNAL_TABLE"

  storage_descriptor {
    location      = "s3://${aws_s3_bucket.source_bucket.id}/matches"
    input_format  = "org.apache.hadoop.mapred.TextInputFormat"
    output_format = "org.apache.hadoop.hive.ql.io.IgnoreKeyTextOutputFormat"
    
    ser_de_info {
      name = "test"
      serialization_library = "org.openx.data.jsonserde.JsonSerDe"
      
      parameters = {
        "serialization.format" = 1
      }
      
    }

    columns {
       name = "version"
       type = "string"
    }

    columns {
       name = "solutionname"
       type = "string"
    }

    columns {
       name = "solutionnumber"
       type = "string"
    }

    columns {
       name = "resourcename"
       type = "string"
    }

    columns {
       name = "resourcetype"
       type = "string"
    }

    columns {
       name = "dependson"
       type = "array<string>"
    }

    columns {
       name = "patternname"
       type = "string"
    }

    columns {
       name = "patterntarget"
       type = "string"
    }

    columns {
       name = "matchespattern"
       type = "boolean"
    }

    columns {
       name = "attributes"
       type = "array<struct<name:string,value:string>>"
    }

  }

}
