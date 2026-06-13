## Governance service

What is a governance service?
It is a service that determines whether to allow you access or not.
If you are a user then you are given a token else given a 401 or whatever is the status code.
Now what we can do is give him a role as well according to that and allow whatever it is we are providing.
Now in the power we can optimize it by using Redis and using different method of auths with different endpoints. there are different ways to role to a user i need to do research on that.


## the key things i want:
- It should handle both *Authentication* and *Authorization*.
- It should be fast.
- It should be cool.

## The Important thing
the main thing is that i need to make this project be divided into small parts then i will be able to make it successfully.
what is a governance microservice?
1. it is just a server that gives a token.
2. make a http server
  - i have made the HTTP server in the previous commit it returns 200 response. i did a bit of refactor but i am thinking of how what to do here right now. do i need more refactor or not i think i should make a package for routes as well and just import that.
3. give the Http server some end points: 
  - this is completed as i made a health checkpoint and necessary refactors now i need to add two things i think logging and rate limitind middleware for now
4. make the endpoints return token
5. store the tokens for verification
6. then we will see.
