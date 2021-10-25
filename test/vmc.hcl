pattern_set_name = "Rehost"

pattern "vmc_rehost" {
  description = "Move a host, as-is from on-prem to cloud using VMC"
  weight      = 99
  target      = "Simple VMC host move + reip"

  rule {
    resource = "server"
    
    condition {
      attribute = "cores"
      operator  = "lt"
      value     = 8
    }

    condition {
      attribute = "os"
      operator  = "eq"
      value     = "Windows"
    }

  }
}

pattern "vmc_loadbalancer" {
  description = "Provide load balancing using AWS ALB from connected VPC"
  weight      = 99
  target      = "AWS EC2 ALB"

  rule {
    resource = "load_balancer"

    condition {
      attribute = "protocol"
      operator  = "eq"
      value     = "HTTPS"
    }
  }
}

pattern "rehost_db_with_nas" {
  description = "Example of a pattern which looks at two resources"
  weight      = 99
  target      = "VMC migration of databases, connect to netapp on-prem"

  rule {
    resource = "nas"

    condition {
      attribute = "type"
      operator  = "eq"
      value     = "netapp"
    }
  }

  rule {
    resource = "database"

    condition {
      attribute = "type"
      operator  = "eq"
      value     = "MSSQL"
    }

    condition {
      attribute = "virtual"
      operator  = "eq"
      value     = "true"
    }
  }
}

pattern "rds_database" {
  description = "Migrate database to AWS RDS"
  weight      = 50
  target      = "AWS RDS database migration"

  rule {
    resource = "database"

    condition {
      attribute = "type"
      operator  = "eq"
      value     = "MSSQL"
    }

    condition {
      attribute = "virtual"
      operator  = "eq"
      value     = "true"
    }
  }
}

