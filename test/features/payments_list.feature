Feature: List payments
  In order to manage payments
  As a product owner
  I need to list existing payments 
    
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
        
