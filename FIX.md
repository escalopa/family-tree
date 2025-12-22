# Fix

## User's page

- Seeing user's profile (changes made) should be visable to users with roles Admin, changing state activite and role can only be done by SuperAdmin
- The users page must be accesable to users with Role Admin, but can't change role or activitiness of other members
- Recent changes made by user's should be clickable to open the change diff between versions

## Mmeber's page

- Adding parent and mother must create a spouse relation if doesn't exist, removing father_id and mother_id shouldn't delete the realtion, the spouse relation is only editable through one of the spouses
- On the memeber card show his fullname in english and arabic, it's deduced from the parent name
