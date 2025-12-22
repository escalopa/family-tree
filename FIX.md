# Fix

## User's page

- Seeing user's profile (changes made) should be visable to users with roles Admin, changing state activite and role can only be done by SuperAdmin
- The users page must be accesable to users with Role Admin, but can't change role or activitiness of other members
- Recent changes made by user's should be clickable to open the change diff between versions

## Mmeber's page

- Adding parent and mother must create a spouse relation if doesn't exist, removing father_id and mother_id shouldn't delete the realtion, the spouse relation is only editable through one of the spouses
- On the memeber card show his fullname in english and arabic, it's deduced from the parent name

## DB

- Merge all migrations to the first db file, I'll recreate the DB

## Auth

- Once I change the role or activitiness of a user he can access the resources automatically, he doesn't need to relogin, so use the information in DB to check access instead from token
- If access_token is missing on the frontend or backend, this isn't enough to count the users as logged out, add an endpoint `/me` that returns the user data, the backend automatically refreshes the tokens if needed, if that isn't possible then 401 is returned

## Other

- Confirm that all methods used by frontend exists on the backend, if some are missing add them
