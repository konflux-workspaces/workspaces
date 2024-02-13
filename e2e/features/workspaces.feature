Feature: Workspace lifecycle

  @fast
  Scenario: user onboarding
    When An user onboards
    Then Default workspace is created for them
    And  The workspace visibility is set to "private"

  @fast
  Scenario: workspace request a private workspace
    Given An user is onboarded
    When  The user requests a new private workspace
    Then  A private workspace is created
  
  @fast
  Scenario: workspace request a community workspace
    Given An user is onboarded
    When  The user requests a new community workspace
    Then  A community workspace is created

  @fast @wip
  Scenario: visibility changes from private to community
    Given A private workspace exists for an user
    When  The owner changes visibility to community
    Then  The workspace is readable for everyone

  @fast
  Scenario: visibility changes from community to private
    Given A community workspace exists for an user
    When  The owner changes visibility to private
    Then  The workspace is readable only for the ones directly granted access to

  @kubespace # @wip
  Scenario: role is set on workspace for owner user 
    When A workspace is created for an user
    Then The owner is granted admin access to the workspace
