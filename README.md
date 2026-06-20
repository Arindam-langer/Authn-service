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
4. make the endpoints return token: i have completed that now an endpoints generate token and header is storing them for the client. we need to do right now is make another token get stored in cookies which is a refresh token
5. store the tokens for verification: completed this but now very good check right now since we need to store the token somewhere for us to check
i think we can two ways here one where we store a raw token in redis cache and just check if it is present in redis if it is and expiry is not happening we just pass it on verified. but we can add another check from our db which stores user data as well and check if that guy is a valid user as well. a two way check would go like this in design.

## Making the check:
login request happens
we provide token back in header
we store the token in redis along with its expiry
and store the claims if user does not exist before using claims in postgres or any db of choice
**now the check**

6. then we will see.
