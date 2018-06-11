job "go-fresh-pr-govendor" {
  datacenters = ["dc1"]

  type = "batch"

  parameterized {
    payload       = "forbidden"
    meta_required = [
      "PROJECT", 
      "GIT_REMOTE", 
      "GIT_BRANCH", 
      "DEPENDENCY", 
      "TOVERSION", 
      "TOREVISION",
    ]
  }

  group "pr" {
    count = 1

    restart {
      attempts = 0
    }

    ephemeral_disk {
      # When sticky is true and the task group is updated, the scheduler
      # will prefer to place the updated allocation on the same node and
      # will migrate the data. This is useful for tasks that store data
      # that should persist across allocation updates.
      # sticky = true
      #
      # Setting migrate to true results in the allocation directory of a
      # sticky allocation directory to be migrated.
      # migrate = true

      # The "size" parameter specifies the size in MB of shared ephemeral disk
      # between tasks in the group.
      size = 1000
    }

    task "pr" {
      driver = "docker"

      config {
        image = "gofrsh/govendor-pr:latest"
        args = [
          "${NOMAD_META_PROJECT}", 
          "${NOMAD_META_GIT_REMOTE}", 
          "${NOMAD_META_GIT_BRANCH}", 
          "${NOMAD_META_DEPENDENCY}", 
          "${NOMAD_META_TOVERSION}", 
          "${NOMAD_META_TOREVISION}",
        ]
      }

      env {
        "GIT_USER_NAME"   = "go-fresh-bot"
        "GIT_USER_EMAIL"  = "email@example.com"
        "GITHUB_USERNAME" = "go-fresh-dev"
        "GITHUB_TOKEN" = "abc"
      }
    }
  }
}
