Feature: Update workspaces via REST API

  Scenario: users can update their workspaces' visibility
    Given An user is onboarded
    And   Default workspace is created for them
    When  The user changes workspace visibility to "community"
    Then  The workspace visibility is updated to "community" 

  Scenario: users can not update visibility of non-owned workspaces
    Given A community workspace exists for an user
    And   User "alice" is onboarded
    Then "alice" can not change workspace visibility to "community"

  Scenario: visibility changes from private to community
    Given A private workspace exists for an user
    When  The owner changes visibility to community
    Then  The workspace is readable for everyone

  @skip
  Scenario: visibility changes from community to private
    Given A community workspace exists for an user
    When  The owner changes visibility to private
    Then  The workspace is readable only for the ones directly granted access to

