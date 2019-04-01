Feature: List payments
  In order to manage payments
  As a product owner
  I need to list existing payments 

  Scenario: Invalid request
    When I get payments without from/to
    Then I should have status code 400
    
  Scenario: No payments
    When I get all payments
    Then I should have status code 200
    And I should have a json
    And that json should have 0 items

  Scenario: A single payment
    Given I created a new payment with id abc
    When I get all payments
    Then I should have status code 200
    And I should have a json
    And that json should have 1 items
    
  Scenario: Default results page
    Given I created 100 payments
    When I get all payments
    Then I should have status code 200
    And I should have a json
    And that json should have 20 items

  Scenario: Get less payments
    Given I created 100 payments
    When I get payments 0 to 10
    Then I should have status code 200
    And I should have a json
    And that json should have 10 items

  Scenario: Try to get more payments
    Given I created 100 payments
    When I get payments 0 to 50
    Then I should have status code 200
    And I should have a json
    And that json should have 20 items

  
  



  
