Feature: Create payments
  In order to manage payments
  As a product owner
  I need to create new payments 

  Scenario: No payments
    Given a payment with id abc
    When I create that payment
    Then I should have status code 201
    And I should have a json
    And that json should have string at data.id equal to abc
    And that json should have int at data.version equal to 0
    And that json should have a data.attributes.amount
    And I should have 1 payment(s)

  Scenario: Existing payment
    Given I created a new payment with id abc
    When I create that payment
    Then I should have status code 409
    And I should have 1 payment(s)
    
  Scenario: Payment with wrong version
    Given a payment with id abc
    And that payment has version 2
    When I create that payment
    Then I should have status code 201
    And I should have a json
    And that json should have string at data.id equal to abc
    And that json should have int at data.version equal to 0
