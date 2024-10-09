Feature: Workspace Cluster URL

  Scenario: Cluster URL is set
    Given An user is onboarded
    And   Default workspace is created for them
    And   Workspace's Space has cluster URL set
    Then  Workspace has cluster URL in status

  Scenario: Cluster URL is not set
    Given An user is onboarded
    And   Default workspace is created for them
    And   Workspace's Space has no cluster URL set
    Then  Workspace has no cluster URL in status
