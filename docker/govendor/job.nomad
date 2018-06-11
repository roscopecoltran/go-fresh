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
        "GITHUB_TOKEN"    = "abc"
      }
    }
  }
}
