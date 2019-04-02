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
    
  Scenario: Payment without an organisation
    Given a payment without organisation, and id abc
    When I create that payment
    Then I should have status code 400
    And I should have 0 payment(s)
    
  @wip
  Scenario: Payment with a negative amount
    Given a payment with id abc and amount -5.00
    When I create that payment
    Then I should have status code 400
    And I should have 0 payment(s)
