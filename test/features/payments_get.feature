Feature: Get a payment
  In order to manage payments
  As a product owner
  I need to be able to fetch individual payments 
    
  Scenario: Non existing payment
    Given a payment with id abc
    When I get that payment
    Then I should have status code 404
    
  Scenario: Existing payment
    Given I created a new payment with id abc
    When I get that payment
    Then I should have status code 200
    
  Scenario: Payment previously deleted
    Given I created a new payment with id abc
    And I deleted that payment
    When I get that payment
    Then I should have status code 404
