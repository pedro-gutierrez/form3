Feature: Delete payments
  In order to manage payments
  As a product owner
  I need to delete  existing payments 
    
  Scenario: Non existing payment
    Given a payment with id abc
    When I delete that payment
    Then I should have status code 404
    
  Scenario: Existing payment
    Given I created a new payment with id abc
    When I delete that payment
    Then I should have status code 204
    And I should have 0 payment(s)

  Scenario: Payment already deleted
    Given I created a new payment with id abc
    And I deleted that payment
    When I delete that payment
    Then I should have status code 404

  @wip
  Scenario: Obsolete version
    Given I created a new payment with id abc
    And I updated that payment
    When I delete version 1 of that payment
    Then I should have status code 409

