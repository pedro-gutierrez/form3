Feature: Health
  In order to deploy this service into a high available cluster
  As a site reliability manager
  I need to query a health check endpoint on my service

  Scenario: Service is up
    When I query the health endpoint
    Then I should have status code 200
    And I should have a json
    And that json should have string at status equal to up
