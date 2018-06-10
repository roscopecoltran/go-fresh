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
      # The number of attempts to run the job within the specified interval.
      attempts = 0
      interval = "30m"

      # The "delay" parameter specifies the duration to wait before restarting
      # a task after it has failed.
      delay = "15s"

     # The "mode" parameter controls what happens when a task has restarted
     # "attempts" times within the interval. "delay" mode delays the next
     # restart until the next interval. "fail" mode does not restart the task
     # if "attempts" has been hit within the interval.
      mode = "fail"
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
      size = 300
    }

    task "pr" {
      driver = "docker"

      config {
        image = "go-fresh-govendor"
        args = [
          "${NOMAD_META_PROJECT}", 
          "${NOMAD_META_GIT_REMOTE}", 
          "${NOMAD_META_GIT_BRANCH}", 
          "${NOMAD_META_DEPENDENCY}", 
          "${NOMAD_META_TOVERSION}", 
          "${NOMAD_META_TOREVISION}",
        ]
      }
    }
  }
}
