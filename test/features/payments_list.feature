Feature: List payments
  In order to manage payments
  As a product owner
  I need to list existing payments 
    
  @wip
  Scenario: No from/to
    When I get payments without from/to
    Then I should have status code 200
    And I should have a json
    And that json should have 0 items

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
    And that json should have a data
    And that json should have a data[0].id
    And that json should have a data[0].type
    And that json should have a data[0].version
    And that json should have a data[0].attributes.amount
    And that json should have a links
    And that json should have a links.self
    And that json should have a links.next

  Scenario: Prev link
    When I get payments 20 to 40
    Then I should have status code 200
    And I should have a json
    And that json should have a links
    And that json should have a links.prev
    And that json should have a links.self
    And that json should have a links.next
    
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

  
  



  
