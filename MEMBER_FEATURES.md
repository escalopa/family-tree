# Member Management Features

## ✅ Spouse Management on Edit Member Page

### How to Delete a Spouse:
1. Open the member's edit dialog
2. Scroll down to the "Spouses" section
3. Each spouse card has two buttons on the right:
   - **Blue Edit Button** (✏️): Edit marriage/divorce dates
   - **Red Delete Button** (🗑️): Remove the spouse relationship
4. Click the **Delete Button**
5. Confirm the deletion in the dialog
6. The spouse relationship will be removed

**Note**: Spouse relationships with children cannot be deleted to maintain data integrity.

### Features:
- ✅ Edit marriage and divorce dates
- ✅ Delete spouse relationships (if no children)
- ✅ Add new spouses
- ✅ Click on spouse cards to navigate to their profile

---

## ✅ Single Parent Support

### You Can Create Members with Only One Parent:

The system fully supports members with only one known parent:

**Examples:**
- Only father is known → Leave mother field empty
- Only mother is known → Leave father field empty
- Both parents unknown → Leave both fields empty
- Both parents known → Fill both fields

### How It Works:
1. When creating/editing a member, the parent fields are **optional**
2. Select a parent from the autocomplete, or leave it blank
3. **Spouse Relationship Auto-Creation:**
   - If BOTH father AND mother are set → System automatically creates a spouse relationship between them (if it doesn't exist)
   - If ONLY one parent is set → NO spouse relationship is created
   - If you later add the second parent → Spouse relationship is created automatically

### Use Cases:
- **Orphan Records**: Member with no known parents
- **Single Parent Families**: Only one parent is known/recorded
- **Incomplete Data**: Waiting to research and add the second parent later
- **Historical Records**: Common in genealogy when only partial information is available

---

## 🎯 Member Dialog Features Summary

### Navigation:
- ✅ Click on **Father card** → Opens father's profile
- ✅ Click on **Mother card** → Opens mother's profile
- ✅ Click on **Spouse card** → Opens spouse's profile
- ✅ Click on **Child card** → Opens child's profile

### Information Displayed:
- ✅ Member's photo
- ✅ Full name (English & Arabic) - automatically computed from lineage
- ✅ Basic information (DOB, DOD, profession, nicknames)
- ✅ Parents (if any)
- ✅ Spouses (if any) with married years
- ✅ Children (if any) sorted by birth date

### Actions:
- ✅ Add/Edit/Delete member
- ✅ Upload/Change/Delete photo
- ✅ Add spouse
- ✅ Edit spouse dates (marriage/divorce)
- ✅ Delete spouse (if no children)
- ✅ Navigate between related members

---

## 🔒 Data Integrity Rules

1. **Circular Relationships**: Prevented (member cannot be their own ancestor)
2. **Spouse with Children**: Cannot delete spouse relationship if children exist with both parents
3. **Gender Validation**: Father must be male, mother must be female in spouse relationships
4. **Version Control**: Optimistic locking prevents concurrent update conflicts

---

## 💡 Tips

- **Navigation**: The member dialog is a powerful navigation tool - click on any family member to explore the tree
- **Auto-Spouse Creation**: When you add both parents to a child, the system automatically creates the marriage relationship if it doesn't exist
- **Incomplete Data**: It's perfectly fine to leave parent fields empty and add them later when you discover more information
- **Photo Management**: Photos are versioned to prevent caching issues - they update immediately after upload
