# Design As Code
A tool to express application designs as code and match those designs to patterns

This is work-in-progress, and there is lots still [to-do](#to-do)

## How it works
This tool allows you to define your application design as code similar to Terraform.  It uses the HCL language, and has a configurable schema, but it comes with some pre-created constructs including:

* load_balancer
* server
* nas
* database

It is designed to bridge the gap between traditional design documentation and a CMDB.  Design documentation suffers because once it is written and used in the implementation process, it is rarely looked at or maintained.  CMBDs suffer because they tend to persist highly technical information about the infrastructure layers, but they rarely capture the semantics of an application.

It is often the case that IT teams need to find all their applications with particular characteristics - especially to service needs like migration projects, where the migration paths are dependent on the characteristics of the applications.

This language is designed to capture those semantics, so that application solutions can be managed as code, and queried and matched based on rules.  The pattern rules are also captured as HCL and can be loaded and used by the design-as-code tool.

## Changing the schema
As mentioned above, the schemas used to describe the resources are configurable.  You can change the `solution-spec.yml` file to change these schemas.

Here's an example:

```yml
database:
  type: string
  platform: string
  arch: string
  virtual: bool
  ha: bool
  role: string
  sla: block

sla:
  availability: string
  rto: string
  rpo: string
```

Each top-level item becomes a 'block' in the expected HCL file, with the attributes defined under it - these attributes are all optional in the resulting HCL schema spec.  Blocks can be nested, and you can see an example of this here with `database` containing `sla`.

The `depends_on` attribute, which is used to specify topological relationships between items will be automatically added to each block, this does not need to be specified.  The schema parsing will fail if `depends_on` is specified as an attribute.

## Example

This is a simple 2-tier app, with a UI tier and a backend database:

```hcl
solution_name = "Richard's test app"
apm_number    = "APM00001"

resource "load_balancer" "lb" {
  protocol = "HTTPS"
  backends = "server.ui"
}

resource "server" "ui" {
  os         = "Windows"
  virtual    = true
  hypervisor = "vmware"
  arch       = "x86"
  cores      = 2
  memory     = 8
  role       = "active"
  count      = 2

  depends_on = [
    database.db,
    nas.cache
  ]
}


resource "nas" "cache" {
  type = "netapp"
}

resource "database" "db" {
  type     = "MSSQL"
  platform = "Windows"
  arch     = "x86"
  virtual  = true
  ha       = true
  role     = "primary"

  sla {
    availability = "5nines"
    RTO          = "1hr"
    RPO          = "5mins"
  }
}
```

## Patterns

Let's imagine we are running a cloud migration project and we want to match our application to a library of cloud migration paths.  Typically we want to break down the application into its underlying components and find appropriate treatment options for each component.  We call those options Patterns, and we can express patterns with rules which can match one or more resources which meet certain expectations.

Here's an example of a pattern which matches Windows servers with less than 8 CPU cores

```hcl
pattern "vmc_rehost" {
  description = "Move a host, as-is from on-prem to cloud using VMC"
  weight      = 99
  target      = "Simple VMC host move"

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
```

## Pattern matching

Patterns can match one or more resources and they can be given arbitary weights.  The process of pattern matching for a given application takes 2 passes:

1) Do an intial match of all the patterns against the application resources, return all the matches, even if they compete for the same resources
2) 'Solve' for a given goal by selecting patterns so that each resouce is claimed by at most one pattern

Two solvers are currently supported:

* Max coverage: selects the patterns which claim the most resources
* Max priority: selects the patterns with the smallest weights, even if that means leaving some resources untreated

### Example output

Here's an example output run against our example 2-tier app, with a slightly bigger rule-set, in this instance it was solving for max-priority.

```
+---+------------------+----------------------------+------------------+
| # | PATTERN          | TARGET                     | RESOURCES        |
+---+------------------+----------------------------+------------------+
| 0 | rds_database     | AWS RDS database migration | database/db      |
| 1 | vmc_rehost       | Simple VMC host move       | server/ui        |
| 2 | vmc_loadbalancer | AWS EC2 ALB                | load_balancer/lb |
+---+------------------+----------------------------+------------------+
```

## Running the tool

1. Download and compile the tool `go build`
2. Create your solution and patterns file, called `app.hcl` and `patterns.hcl` respectively
3. Run the tool `./design-as-code`

The tool defaults to looking for app.hcl and patterns.hcl, it also defaults to the priority solver.  You can change these defaults as follows

```
Usage of ./design-as-code:
  -app string
        Path to the file containing the list of patterns to use for matching. (default "app.hcl")
  -debug
        Should we log verbose messages for debugging?
  -patternlib string
        Path to the file containing the list of patterns to use for matching. (default "patterns.hcl")
  -solvefor string
        What solution mode should we use. (default "priority")
```

## TO-DO

There is still much to do, current goals:

1. Introduce a structured output e.g. JSON as well as the tabular output
2. Extend the rule language to include complex conditionals and additional operators
3. Support rules which use solution topology (e.g. if this is linked to that then...)







