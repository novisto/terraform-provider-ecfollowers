terraform {
  required_providers {
    ecfollowers = {
      source = "novisto/ecfollowers"
    }
  }
  required_version = ">= 1.0.3"
}

provider "ecfollowers" {
  url = "changeme"
  username = "changeme"
  password = "changeme"
}

resource "ecfollowers_follower_index" "my_data" {
  index = "data-from-far-away"
  remote_cluster = "my-cluster"
  leader_index = "data"
}
