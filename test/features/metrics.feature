Feature: Metrics 
  In order to measure performance
  As a site reliability engineer
  I need to obtain real time operational metrics

  Scenario: Cold start
    When I query the metrics endpoint
    Then I should have status code 200
    And I should have content-type text/plain 
    And I should have a text
    And that text should match go_goroutines
