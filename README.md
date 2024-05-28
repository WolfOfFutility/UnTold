# Untold
A simple OpenSource No-SQL database to empower local development of developer projects. All objects are saved to file for re-use in later iterations.

This repository currently only uses base go modules - no third party open source options are used. 

## 1.0 - Queries
The store can be queried with a few basic calls. Each of these calls fall into one of the C.R.U.D. categories and are structured similarly to the examples below.

- PUSH
    - Creates a new row within a store "table" with values provided.
    - Examples:
        - ``` PUSH Username = Admin, Password = admin TO Users ```

- PULL
    - Returns row object(s) based on query parameters.
    - Examples:

- PUT
    - Updates row object(s) based on query parameters.
    - Examples:

- DELETE
    - Removes row object(s) based on query parameters.
    - Examples:

## 2.0 - Encryption
The database is protected by two different types of encryption; Symmetric and Asymmetric encryption.

### 2.1 - Symmetric Encryption 
Symmetric Encryption is applied over all .dat files, which is locked by the main.dat key. *Keep this key safe, this provides access to usernames, passwords and private keys, which could be used for iterating other secrets.*

### 2.2 - Asymmetric Encryption
Asymmetric Encryption (Public Key / Private Key) is used to protect secrets for individual users. A private key is stored in each user profile, which is then used to generate a public key for users as an Auth Token. Whenever a user completes an action, the auth token is validated against another public key generated by the user's private key. Each of the user's secrets are encrypted with the public key, and can only be decrypted with their private key.

## 3.0 - Role-Based Access Control (RBAC)
The database has layers of Role-Based Access Control added for more security around the data. This is split up into multiple concepts, including Users, Roles, Groups and Policies.

### 3.1 - Users
Users provide individuals with scopeable access to each of the databases through a username and password. Once logged in, the user is sent back a public version of their login, to scope down data as much as possible. 

### 3.2 - Groups
To simplify management of users and their related access, groups exist to create a logical collection of users. Groups can be assigned to roles. A key example would be to create a group for a team, and provide them with all the same access. 

### 3.3 - Roles
Roles serve as an easy to use medium to provide access to users and groups to a specific scope. Roles can be created, or default roles used for the management of each of the databases. By default, Root Admin, Root Writer and Root Reader are created on Database initialisation. 

### 3.4 - Policies
Policies serve as a way to communicate the actual permissions being provided within a role. Examples of a policy might be a Reader policy that allows ```PULL``` queries. Scoping is provided at the Role level, policies exist only for declaritive allowance of actions.

## Coming Soon
- Access Reviews
- Application Context RBAC
- Internal and external MFA integrations
- Mermaid diagrams and robust documentation
- More stable query structures
- More complex query structures, including creation, deletion and joining of store tables
- Fast store filling, allowing for test data to be rapidly created
- Dynamic key source for Data Encryption
- Dynamic Data Masking
- Multi-Store Replication
- Data Transferrence to SQL and No-SQL Formats and Databases
- Vector Database Mode
- Matrix Database Mode
- Integration Modules
    - PowerShell
    - Python
    - JavaScript / TypeScript