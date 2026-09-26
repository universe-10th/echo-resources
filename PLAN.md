- For each file not having tests, please review them and make tests.
- Run the tests in Golang in the whole page.
- Make an overall test for the platform, involving:
  - Several levels
  - Mixing singleton and collection results
  - Mixing soft-deleted and hard-deleted resources
  - Create another Storage engine, intended as MEMORY. Make it available
    for users (like Gorm and MongoDB), and use it in this overall test.
  - Make it for Echo.