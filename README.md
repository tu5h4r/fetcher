# Fetcher
Library based on goLang for HTML parsing and retrieving particular dataset
<br>
In its 0.01 version, the code simply parses an HTML page retries the questions, how many votes, comments,etc and prints out in a File/Console.
<br> Pretty Basic.<br>This is what the result of parsed file looks like currently

```javascript
Votes: 4
Comments: 33
URL: https://www.careercup.com/question?id=5749533368647680
ID: 5749533368647680
Question:
Given the root of a binary tree containing integers, print the columns of the tree in order with the nodes in each column printed top-to-bottom.
Input:
      6
     / \
    3   4
   / \   \
  5   1   0
 / \     /
9   2   8
     \
      7

Output:
9 5 3 2 6 1 7 4 8 0

Input:
       1
     /   \
    2     3
   / \   / \
  4   5 6   7

When two nodes share the same position (e.g. 5 and 6), they may be printed in either order:

Output:
4 2 1 5 6 3 7
or:
4 2 1 6 5 3 7

```
