# API

## Logic

### Tree

* Show tree by root (default highest root)  
* Show tree by 2 members

### Scoring

* Mandatory  
  * Arabic Name - 1p  
  * English Name - 1p  
  * Gender - 1p  
* Optional:  
  * Picture - 2p  
  * Date of birth - 5p  
  * Date of death - 5p  
  * reference to parent father node - 3p  
  * reference to parent mother node - 3p  
  * reference to spouse node - 3p  
  * List of nicknames - 1p (for all names)  
  * Profession 1p

### Generation level

* Calculated using this formula

```python
max(father_id+1,mother_id+1,max(spouse_id))
```

### Image

* Member's photo is returned by an unique endpoint (excepts many member_id)
* Max Image size 3mb
* Image is set using unique handler (not in member post handler)
* Support ONLY popular image types

### Recent activities

* Have paginations

### Role

* None (default for any new user)
* Viewer
* Admin  
* SuperAdmin

### Rollback

* On rollback create a new version
* Retore to the state of the specified

### Changes

* All changes to members fields must support version_id to resolve conflicts

### Spouse

* On spouse change write changes to members_history
* In members_history also change in spouses are recorded
* marriage_date must be before divorce_date

### Authentication

* auth_token, refresh_token and session_id are stored in cookie
* On auth middleware if auth_token is old try to renew with refresh_token

### Roles

Created manually using SQL by administrator

### Users

* Created only using google oauth
* full_name, email, avatar are set from oauth response, on email conflict must update full_name and avatar
* is_active can be changed using `PUT users/USER_ID/active` which has boolean flag of the value to set
* User role can be promoted and demoted using `PUT users/USER_ID/role`
* list users (no pagination)
* logout, (logout from all devices), admin_logout(called by super_admin to logout )

### Members (members_spouse)

* Create, logged to history
* Update, all changes are logged in into history
* Picture are added/removed using a unique handler (not post,put,patch of member)
* Delete, set values to nothing or default value
* On create, update, update users scores column in users and user_scores table

## Stack

### Go

* Gin
* Clean architecture
* pgx
* Docker,Dockerfile
