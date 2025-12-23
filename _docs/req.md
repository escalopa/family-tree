# Requirements

## System Requirements

* **REQ-001**: The system shall allow users to register using gmail (oauth2).
  * New user has user role (access to nothing), super admin can promote them
* **REQ-002**: The system shall be accessible on mobile, tablet, and desktop devices.
* **REQ-003**: The system shall encrypt all sensitive data in transit and at rest.
* **REQ-004**: The system shall supprot manipulating memebers (create/update-put,post).
* **REQ-005**: The system shall display a user dashboard with recent activity.
  * Recent members changes by him
  * UserInfo (avatar, name, email, etc.)
  * `/users/{user_id}`
  * Leaderboards
  * `/users/leaderboard`
  * Recent
    * Score Score history
      * `/users/score/{user_id}`
    * Members history
      * `/users/members` - changes made by user

---

## Functional Requirements

* **REQ-007**: family member node shall have the following properties
  * Mandatory:
    * Arabic Name
    * English Name
    * Gender
  * Optional:
    * Picture
    * Date of birth
    * Date of death
    * reference to parent father node
    * reference to parent mother node
    * reference to spouse node
    * List of nicknames
    * Profession
  * Deduced properties:
    * Arabic Full Name
    * English Full Name
    * Age
    * Generation level (relative to root)
* **REQ-008**: family tree shall be viewed in a tree diagram
* **REQ-009**: tree diagram can be started by selecting any desired node as root (if no root is selected select the highest root by age)
  * `/tree?root={member_id}`
* **REQ-010**: family tree view shall support view as list diagram sorted by age/birth date as well
  * `/tree?root={member_id}&style=list`
* **REQ-011**: view can be filtered based on family member properties
  * arabic_name (prefix based)
  * english_name (prefix based)
  * gender
  * married status (yes, no)
  * no pagination
  * filters are applied using AND operation
  * At least one filter must be used
  * `/members/search?arabic_name=ب&english_name=d&gender=M,married=1`
* **REQ-012**: family members shall be searchable based on family member properties
  * arabic_name (prefix based)
  * english_name (prefix based)
  * gender
  * married status (yes, no)
  * no pagination
  * filters are applied using AND operation
* **REQ-013**: it shall be possible to show the relation between 2 given family members in tree diagram
  * `/tree/relation?member1={member1_id}&member2={member2_id}`
* **REQ-014**: node color shall differ based on gender
  * Cyan: man
  * Pink: woman
* **REQ-015**: node name must always be shown
  * In tree view show
    * Name (arabic, english)
    * Picture (if not set user default photo based on gender)
* **REQ-017**: on node click, node can expand to shown all properties
  * Expand is made on a side window on the right
  * `GET /members/info/{member_id}`
* **REQ-018**: each node (either male or female) will have a black colored line link to its children nodes
* **REQ-019**: in case of having a node & its spouse node under the current root node:
  * both nodes will have a pink colored line link to each other
  * children nodes will branch from the father node only
* **REQ-020**: nodes edit history shall be recorded & blamed to users (can rollback to a specific version)
  * `GET /members/history?member_id={member_id}`
* **REQ-021**: there shall be a scoring mechanism to count users' contribution in building the family tree

---

## User Privileges Requirements

* **REQ-022**: privileges are splitted into the following levels where higher levels have access to all the lower level privileges:
  * Super Admin
    * grant admin access to users
      * `PUT /users/role`
    * view of female family members' age/birth dates (restricted to super admins only)
    * disable users
  * Admin
    * add new family member nodes
    * edit existing family member nodes
    * view of female family members' picture (restricted to admins or super admins only)
  * User
    * view family tree
  * None
    * access to nothing until admin prompt him

---

## Future Requirments

* **REQ-006**: The system shall support both Arabic & English languages.
* **REQ-016**: node properties other than name can be optionally shown
