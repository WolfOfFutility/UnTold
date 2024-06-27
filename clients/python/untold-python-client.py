import math
import socket
import json
import base64
import random
import string

BASE_HOST = "127.0.0.1"
BASE_PORT = 8080

## User object to handle authentication
class UserObj:
    Username: str
    PublicToken: str

    def __init__(self, Username: str, PublicToken: str):
        self.Username = Username
        self.PublicToken = PublicToken

## object to handle server requests to and from the server
class GeneralServerRequest:
    RequestType: str
    Payload: dict
    User: UserObj

    def __init__(self, RequestType: str, Payload: str, User: UserObj) :
        self.RequestType = RequestType
        self.Payload = Payload
        self.User = User

    ## handle the encryption of the data to send to the server
    def encryptData() :
        return
    
    def convertToJSON(self):
        json_dump = json.dumps(
            self,
            default=lambda o: o.__dict__, 
            sort_keys=True,
            indent=4)
        
        return bytearray(json_dump, "utf-8")

## untold class to present the module management
class Untold :
    BASE_HOST: str
    BASE_PORT: int
    User: UserObj

    def __init__(self, BASE_HOST: str, BASE_PORT: int):
        self.BASE_HOST = BASE_HOST
        self.BASE_PORT = BASE_PORT
        self.User = None

    ## Handle sending messages to the server
    ## BASE_HOST and BASE_PORT identify the address and port of the server to send requests to
    def send_to_server(self, req: GeneralServerRequest) -> bytes:
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            
            try :
                s.connect((self.BASE_HOST, self.BASE_PORT))
                s.sendall(req.convertToJSON())
                data = s.recv(51200) ## cap out at 50MB received
                
                return data

            except:
                print("Failed to send data to the server")
                return None
            
            finally:
                ## close the connection after everything is done
                s.close()
    
    ## login as a user with the given username and password
    ## the resulting login object returned from the server will be stored against the object in memory
    def login(self, username: str, password: str):
        login_result = self.send_to_server(GeneralServerRequest(
            "login",
            {
                "username": username,
                "password": password
            },
            UserObj("init", str(base64.b64encode(bytearray("init", "utf-8"))))
        ))

        if login_result != None :
            try :
                parsed_login = json.loads(login_result)
                self.User = UserObj(parsed_login["Username"], parsed_login["PublicToken"])

            except Exception as e:
                print(e, login_result)
                return None
    
    ## generate a default table schema from a class, making it easy for object-based ingestion
    def generate_db_table_schema(self, model: object) -> list[dict]:
        try :
            schemaList = []

            for key in model.__annotations__:
                columnName = key
                
                match model.__annotations__[key]:
                    case __builtins__.str:
                        columnType = "string"
                    case __builtins__.int:
                        columnType = "int"
                    case __builtins__.bool:
                        columnType = "bool"
                    case _:
                        columnType = None
                        raise Exception("unrecognised type in object")
                    
                schemaList.append({
                    "ColumnName": columnName,
                    "ColumnType": columnType,
                    "Nullable": False
                })
            
            return schemaList
        
        except Exception as e:
            print(e)
            return None

    ## create a database table, this will automatically use the user as a basis for authentication
    def create_db_table(self, table_name: str, schema: list[dict], primary_key_name: str):
        try:
            if self.User == None :
                raise Exception("no authenticated user object could be found, please run the login command")
            
            create_table_result = self.send_to_server(GeneralServerRequest(
                "push_table",
                {
                    "tableName": table_name,
                    "schema": schema,
                    "primaryKeyName": primary_key_name
                },
                self.User
            ))
            
            print(create_table_result)

        except Exception as e:
            print(e)
    
    ## add table rows to a specified table
    def add_table_rows(self, tableName: str, rowValues: list[dict]):
        try:
            if self.User == None :
                raise Exception("no authenticated user object could be found, please run the login command")
            
            ## handle multiple rows being pushed
            if type(rowValues).__name__ == "list":
                create_row_result = self.send_to_server(GeneralServerRequest(
                    "push_row_multi",
                    {
                        "tableName": tableName,
                        "rowValues": rowValues
                    },
                    self.User
                ))

            ## handle an individual row being pushed
            else :
                create_row_result = self.send_to_server(GeneralServerRequest(
                    "push_row",
                    {
                        "tableName": tableName,
                        "rowValue": rowValues
                    },
                    self.User
                ))
            
            print(create_row_result)
        
        except Exception as e:
            print(e)
    
    ## gets the values of table rows based on a query
    def get_table_row(self, tableName: str, queryString: str):
        try:
            if self.User == None :
                raise Exception("no authenticated user object could be found, please run the login command")
            
            get_row_result = self.send_to_server(GeneralServerRequest(
                "pull_row",
                {
                    "tableName": tableName,
                    "queryString": queryString
                },
                self.User
            ))

            print(get_row_result)
        
        except Exception as e:
            print(e)
    
    ## updates table rows based on query strings
    def update_table_row(self, tableName: str, queryString: str):
        try:
            if self.User == None :
                raise Exception("no authenticated user object could be found, please run the login command")
            
            update_row_result = self.send_to_server(GeneralServerRequest(
                "put_row",
                {
                    "tableName": tableName,
                    "queryString": queryString
                },
                self.User
            ))

            print(update_row_result)
        
        except Exception as e:
            print(e)
    
    ## deletes table rows based on query string
    def delete_table_row(self, tableName: str, queryString: str):
        try:
            if self.User == None :
                raise Exception("no authenticated user object could be found, please run the login command")
            
            delete_row_result = self.send_to_server(GeneralServerRequest(
                "delete_row",
                {
                    "tableName": tableName,
                    "queryString": queryString
                },
                self.User
            ))

            print(delete_row_result)
        
        except Exception as e:
            print(e)
    
    ## create a user
    def create_user(self, username: str, password: str):
        try:
            if self.User == None :
                raise Exception("no authenticated user object could be found, please run the login command")
            
            result = self.send_to_server(GeneralServerRequest(
                "push_user",
                {
                    "username": username,
                    "password": password
                },
                self.User
            ))

            print(result)
        
        except Exception as e:
            print(e)
    
    ## create a group
    def create_group(self, groupName: str):
        try:
            if self.User == None :
                raise Exception("no authenticated user object could be found, please run the login command")
            
            result = self.send_to_server(GeneralServerRequest(
                "push_group",
                {
                    "groupName": groupName
                },
                self.User
            ))

            print(result)
        
        except Exception as e:
            print(e)
    
    ## create a role
    def create_role(self, roleName: str, scope: str, perms: list[str]):
        try:
            if self.User == None :
                raise Exception("no authenticated user object could be found, please run the login command")
            
            result = self.send_to_server(GeneralServerRequest(
                "push_role",
                {
                    "roleName": roleName,
                    "scope": scope,
                    "permissions": perms
                },
                self.User
            ))

            print(result)
        
        except Exception as e:
            print(e)
    
    ## find a group by its name
    def find_group(self, groupName: str) -> dict:
        try:
            if self.User == None :
                raise Exception("no authenticated user object could be found, please run the login command")
            
            result = self.send_to_server(GeneralServerRequest(
                "pull_group",
                {
                    "groupName": groupName
                },
                self.User
            ))

            return (json.loads(result))
        
        except Exception as e:
            print(e)
    
    ## find a role by its name
    def find_role(self, roleName: str) -> dict:
        try:
            if self.User == None :
                raise Exception("no authenticated user object could be found, please run the login command")
            
            result = self.send_to_server(GeneralServerRequest(
                "pull_role",
                {
                    "roleName": roleName
                },
                self.User
            ))
            
            return (json.loads(result))
        
        except Exception as e:
            print(e)

    ## assign a group to a role
    def assign_group_to_role(self, groupID: int, roleID: int): 
        try:
            if self.User == None :
                raise Exception("no authenticated user object could be found, please run the login command")
            
            result = self.send_to_server(GeneralServerRequest(
                "push_group_role_assignment",
                {
                    "groupID": groupID,
                    "roleID": roleID
                },
                self.User
            ))

            print(result)
        
        except Exception as e:
            print(e)
    
    ## assign a user to a role
    def assign_user_to_role(self, username: str, roleID: int):
        try:
            if self.User == None :
                raise Exception("no authenticated user object could be found, please run the login command")
            
            result = self.send_to_server(GeneralServerRequest(
                "push_user_role_assignment",
                {
                    "username": username,
                    "roleID": roleID
                },
                self.User
            ))

            print(result)
        
        except Exception as e:
            print(e)

    ## assign a user to a group
    def assign_user_to_group(self, username: str, groupID: int):
        try:
            if self.User == None :
                raise Exception("no authenticated user object could be found, please run the login command")
            
            result = self.send_to_server(GeneralServerRequest(
                "push_user_group_assignment",
                {
                    "username": username,
                    "groupID": groupID
                },
                self.User
            ))

            print(result)
        
        except Exception as e:
            print(e)
    
    ## remove a user from a group
    def remove_user_from_group(self, username: str, groupID: int):
        try:
            if self.User == None :
                raise Exception("no authenticated user object could be found, please run the login command")
            
            result = self.send_to_server(GeneralServerRequest(
                "delete_user_group_assignment",
                {
                    "username": username,
                    "groupID": groupID
                },
                self.User
            ))

            print(result)
        
        except Exception as e:
            print(e)

    ## remove a user from a role
    def remove_user_from_role(self, username: str, roleID: int):
        try:
            if self.User == None :
                raise Exception("no authenticated user object could be found, please run the login command")
            
            result = self.send_to_server(GeneralServerRequest(
                "delete_user_role_assignment",
                {
                    "username": username,
                    "roleID": roleID
                },
                self.User
            ))

            print(result)
        
        except Exception as e:
            print(e)
    
    ## remove a group from a role
    def remove_group_from_role(self, roleID: int, groupID: int):
        try:
            if self.User == None :
                raise Exception("no authenticated user object could be found, please run the login command")
            
            result = self.send_to_server(GeneralServerRequest(
                "delete_group_role_assignment",
                {
                    "groupID": groupID,
                    "roleID": roleID
                },
                self.User
            ))

            print(result)
        
        except Exception as e:
            print(e)
    
    ## delete a user
    def delete_user(self, username: str):
        try:
            if self.User == None :
                raise Exception("no authenticated user object could be found, please run the login command")
            
            result = self.send_to_server(GeneralServerRequest(
                "delete_user",
                {
                    "username": username
                },
                self.User
            ))

            print(result)
        
        except Exception as e:
            print(e)

    ## delete a group
    def delete_group(self, groupID: int):
        try:
            if self.User == None :
                raise Exception("no authenticated user object could be found, please run the login command")
            
            result = self.send_to_server(GeneralServerRequest(
                "delete_group",
                {
                    "groupID": groupID
                },
                self.User
            ))

            print(result)
        
        except Exception as e:
            print(e)
    
    ## delete a role
    def delete_role(self, roleID: int):
        try:
            if self.User == None :
                raise Exception("no authenticated user object could be found, please run the login command")
            
            result = self.send_to_server(GeneralServerRequest(
                "delete_user",
                {
                    "roleID": roleID
                },
                self.User
            ))

            print(result)
        
        except Exception as e:
            print(e)

    ## delete a table
    def delete_table(self, tableName: str):
        try:
            if self.User == None :
                raise Exception("no authenticated user object could be found, please run the login command")
            
            result = self.send_to_server(GeneralServerRequest(
                "delete_table",
                {
                    "tableName": tableName
                },
                self.User
            ))

            print(result)
        
        except Exception as e:
            print(e)

# untold = Untold(BASE_HOST, BASE_PORT)
# untold.login("admin1", "admin")