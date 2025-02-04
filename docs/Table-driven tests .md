## Table-driven tests 

And this of course copying and pasting the whole blocks of code like this really not a good idea, but I'm just trying to make it eminently readable for you. What I would do once I figured out how I'm writing this test is make this a table test and just build up my data as a slice of whatever format I need and run it through a for loop.

But this works just fine for our purposes today, and we're trying to make it clear how everything is getting tested.

*****

A table-driven test is a testing pattern commonly used in programming, especially in languages like Go, where you define multiple test cases in a single data structure (usually a slice or array). Each test case in the table represents an input and the expected output. Then, you iterate through the table using a loop, running the same logic for each test case.

This approach makes your tests more concise, readable, and easier to maintain, especially when you have multiple scenarios to test for the same functionality.


