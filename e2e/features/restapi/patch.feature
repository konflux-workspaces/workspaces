Feature: Patch workspaces via REST API

  Scenario: users can update their workspaces' visibility
    Given An user is onboarded
    And   Default workspace is created for them
    When  The user patches workspace visibility to "community"
    Then  The workspace visibility is updated to "community" 

  Scenario: users can not patch visibility of non-owned workspaces
    Given A community workspace exists for an user
    And   User "alice" is onboarded
    Then "alice" can not patch workspace visibility to "community"
