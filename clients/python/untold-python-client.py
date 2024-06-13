import socket
import json
import base64
import random
# import rsa

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
                data = s.recv(1024)
                
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
    
    ## add a table row to a sepcified table
    ## ** need to find a way to buffer this connection to send heaps at once
    def add_table_row(self, tableName: str, rowValue: dict):
        try:
            if self.User == None :
                raise Exception("no authenticated user object could be found, please run the login command")
            
            create_row_result = self.send_to_server(GeneralServerRequest(
                "push_row",
                {
                    "tableName": tableName,
                    "rowValue": rowValue
                },
                self.User
            ))

            print(create_row_result)
        
        except Exception as e:
            print(e)
    
    ## gets the values of table rows based on a query
    ## ** Need to find a way to buffer this connection to recieve heaps at once
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

class Snake:
    name: str
    colour: str
    breed: str
    age: int

    def __init__(self, name: str, colour: str, breed: str, age: int):
        self.name = name
        self.colour = colour
        self.breed = breed
        self.age = age


untold = Untold(BASE_HOST, BASE_PORT)
# untold.login("python-user", "hiss-hiss")

untold.login("admin1", "admin")

# untold.create_role("python_role", "snakes", ["PULL"])
# untold.create_group("python-group")

# untold.create_user("python-user", "hiss-hiss")

# random_names = [
#     "Della Greer",
#     "Koda Montgomery",
#     "Evangeline Shah",
#     "Zain Mitchell",
#     "Willow Erickson",
#     "Johnny Haley",
#     "Addilynn Hodges",
#     "Alonzo McLean",
#     "Sky Barajas",
#     "Brennan Shaffer",
#     "Alanna York",
#     "Leandro Enriquez",
#     "Nellie Estes",
#     "Hakeem O’Connor",
#     "Charli Frank",
#     "Braylen Hahn",
#     "Fallon Grant",
#     "Leon McIntyre",
#     "Rebekah Stephenson",
#     "Joe Marin",
#     "Celia Olson",
#     "Malachi Salinas",
#     "Royalty Pennington",
#     "Bobby Nixon",
#     "Deborah Blake",
#     "Zyaire Crosby",
#     "Keily Miles",
#     "Jared Stout",
#     "Chana Stephenson",
#     "Joe Hansen",
#     "Hope Floyd",
#     "Pierce Torres",
#     "Violet Hendricks",
#     "Dash Torres",
#     "Violet Gallegos",
#     "Jonas Rocha",
#     "Emmie Jimenez",
#     "Silas Olsen",
#     "Oaklyn Blevins",
#     "Avi Brennan",
#     "Elodie Hampton",
#     "Hank Avalos",
#     "Paloma Bradley",
#     "Richard Cantu",
#     "Galilea Harrell",
#     "Nelson Sandoval",
#     "Elsie Krueger",
#     "Jones Alvarado",
#     "Blake Saunders",
#     "Kasen Beard",
#     "Ezra Rowland",
#     "Eliezer Boyle",
#     "Aliya McCarthy",
#     "Devin McFarland",
#     "Annika Cervantes",
#     "Kamari Cochran",
#     "Alma Kirby",
#     "Tony Brandt",
#     "Loretta Randall",
#     "Trenton Franklin",
#     "Angela Guevara",
#     "Tommy Donaldson",
#     "Natasha Yates",
#     "Braylon Lowery",
#     "Estelle Leon",
#     "Marshall Logan",
#     "Kora McLaughlin",
#     "Ibrahim Melton",
#     "Kamiyah Horton",
#     "Garrett Lyons",
#     "Kenzie Benjamin",
#     "Kyro Felix",
#     "Paisleigh Wiggins",
#     "Azariah McLean",
#     "Sky Frazier",
#     "Callum Campbell",
#     "Addison Sweeney",
#     "Nixon Branch",
#     "Luisa Henry",
#     "Carlos Goodwin",
#     "Shiloh Woods",
#     "Zion Portillo",
#     "Nathalie Morales",
#     "Aaron McCann",
#     "Joyce McConnell",
#     "London Chambers",
#     "Makayla Gaines",
#     "Talon Arroyo",
#     "Kyra Franklin",
#     "Simon Skinner",
#     "Mara Dickerson",
#     "Flynn Maxwell",
#     "Kyla Mullen",
#     "Shepard Hernandez",
#     "Camila Malone",
#     "Ruben McCann",
#     "Joyce Bryant",
#     "Jonah Baker",
#     "Isla Fitzgerald",
#     "Peyton Barr",
#     "Noemi Gray",
#     "Nicholas Delgado",
#     "Alani Norris",
#     "Cairo Eaton",
#     "Miley Landry",
#     "Jaxx Chase",
#     "Angie Shepherd",
#     "Ronald Parks",
#     "Ainsley Cervantes",
#     "Kamari Klein",
#     "Elianna Rios",
#     "Israel Fleming",
#     "Fatima Townsend",
#     "Alexis Suarez",
#     "Jimena Alfaro",
#     "Xzavier Serrano",
#     "Allie Edwards",
#     "Adrian Black",
#     "Molly Cannon",
#     "Archie Hobbs",
#     "Lacey Jefferson",
#     "Raylan Bullock",
#     "Winnie Krueger",
#     "Jones Porter",
#     "Ryleigh Barber",
#     "Solomon Villa",
#     "Johanna Henry",
#     "Carlos Woodard",
#     "Aubrie Villegas",
#     "Clyde Vaughn",
#     "Reign Shannon",
#     "Eliel Cortes",
#     "Lea May",
#     "Finley Quintero",
#     "Keyla Rush",
#     "Kaiser Henson",
#     "Kinslee Roy",
#     "Marcelo Mercado",
#     "Mckinley Dominguez",
#     "Kaden Oliver",
#     "Camille Adams",
#     "Hudson Townsend",
#     "Azalea Zamora",
#     "Quentin Hughes",
#     "Samantha Newton",
#     "Santino Flores",
#     "Emilia Wilkins",
#     "Yusuf Bates",
#     "Madilyn Pace",
#     "Dior Richardson",
#     "Allison Wilcox",
#     "Jerry McGee",
#     "Kayleigh Christian",
#     "Ledger Spears",
#     "Isabela Sullivan",
#     "Evan Hendricks",
#     "Dani Clarke",
#     "Stetson Mosley",
#     "Zaniyah Perez",
#     "Owen Riley",
#     "Kayla Patterson",
#     "Amir Green",
#     "Zoe O’Donnell",
#     "Lian Phelps",
#     "Laney Kramer",
#     "Kylan Vaughn",
#     "Reign Bravo",
#     "Genesis Whitehead",
#     "Sylvie Mullins",
#     "Allen Lester",
#     "Averi Mitchell",
#     "Jaxon Holland",
#     "Mariah Shepard",
#     "Damari Hanna",
#     "Cynthia Foley",
#     "Mohammad Estes",
#     "Brittany Fernandez",
#     "Bentley Gilbert",
#     "Jocelyn Felix",
#     "Rodney Moses",
#     "Karter Hoffman",
#     "Steven Hubbard",
#     "Rosie Nichols",
#     "Atlas Calhoun",
#     "Thalia West",
#     "Diego Rodriguez",
#     "Evelyn McCormick",
#     "Jasiah Roman",
#     "Astrid Rivers",
#     "Bear Mosley",
#     "Zaniyah Watts",
#     "Dakota Fields",
#     "Annie Cantu",
#     "Anakin Brady",
#     "Ryan Roberts",
#     "Josiah Bradley",
#     "Vanessa Levy",
#     "Harold Estrada",
#     "Sawyer Tate",
#     "Dalton Farley",
#     "Wrenley Reese",
#     "Alijah O’Connor",
#     "Charli Nunez",
#     "Caden Rosario",
#     "Louisa Decker",
#     "Taylor McGuire",
#     "April Delarosa",
#     "Osiris Thornton",
#     "Haisley Rasmussen",
#     "Will Blackwell",
#     "Saoirse Brown",
#     "Elijah Whitney",
#     "Madalynn O’brien",
#     "Riley Vega",
#     "Dakota Houston",
#     "Sylas Carlson",
#     "Kali Duffy",
#     "Kyng Smith",
#     "Olivia Gomez",
#     "Isaiah Winters",
#     "Kataleya Cortez",
#     "Zayn Conner",
#     "Alondra Best",
#     "Harlem Garza",
#     "River Esparza",
#     "Carl Espinosa",
#     "Braylee Reynolds",
#     "Vincent Blevins",
#     "Aila Hicks",
#     "Maddox Lester",
#     "Averi Davidson",
#     "Dante Shaffer",
#     "Alanna O’Connell",
#     "Jovanni Lindsey",
#     "Colette Woodard",
#     "Westley Montes",
#     "Roselyn Mayer",
#     "Yahir Greene",
#     "Selena Doyle",
#     "Kashton Robbins",
#     "Stevie Clark",
#     "John Patterson",
#     "Kaylee Avery",
#     "Jakari Ross",
#     "Peyton Salinas",
#     "Edgar Blankenship",
#     "Rosalee Medrano",
#     "Arian Cabrera",
#     "Daleyza Pratt",
#     "Rowen Watson",
#     "Hailey Cameron",
#     "Rayan Knight",
#     "Gracie Frost",
#     "Dario Mason",
#     "Sienna Singh",
#     "Louis Quintana",
#     "Kenia McClain",
#     "Mitchell Harper",
#     "Ana Yu",
#     "Bryant Aguirre",
#     "Ariah Hawkins",
#     "Victor Hammond",
#     "Holly Guerrero",
#     "Bryce Casey",
#     "Sylvia Yang",
#     "Malcolm Short",
#     "Cheyenne Vincent",
#     "Aarav Clayton",
#     "Saige Hancock",
#     "Rex Blevins",
#     "Aila Smith",
#     "Liam Rubio",
#     "Hadassah McKinney",
#     "Romeo Arnold",
#     "Finley Brooks",
#     "Jordan Cantrell",
#     "Yamileth Peck",
#     "Yousef Horn",
#     "Avah Blake",
#     "Zyaire Webster",
#     "Kensley Torres",
#     "Jayden Wiley",
#     "Lauryn Ayers",
#     "Ulises Blair",
#     "Frances Bean",
#     "Mccoy Bryan",
#     "Meredith Ventura",
#     "Branson Truong",
#     "Judith Perry",
#     "Waylon Watkins",
#     "Lola Caldwell",
#     "Rylan Michael",
#     "Aubriella Gilbert",
#     "Tobias Peralta",
#     "Malayah Xiong",
#     "Azrael Powers",
#     "Michelle Torres",
#     "Jayden Wilkins",
#     "Amalia Beck",
#     "Eduardo Franklin",
#     "Angela McBride",
#     "Denver Corona",
#     "Marianna Williams",
#     "Oliver Hawkins",
#     "Ariel Hebert",
#     "Guillermo Melton",
#     "Kamiyah Reyes",
#     "Eli Chase",
#     "Angie Marin",
#     "Aldo Lucero",
#     "Ila Taylor",
#     "Jackson Haley",
#     "Addilynn Gibson",
#     "Tyler Fry",
#     "Clarissa Ross",
#     "Wesley Costa",
#     "Robin Crosby",
#     "Tristen Weeks",
#     "Karen Stevens",
#     "Zachary Richard",
#     "Davina Farley",
#     "Graysen Stephenson",
#     "Khaleesi Reeves",
#     "Clark Reyes",
#     "Audrey Shelton",
#     "Leonel Gonzalez",
#     "Abigail Alvarez",
#     "Xavier Navarro",
#     "Winter Blackburn",
#     "Zahir Ellis",
#     "Ayla Bullock",
#     "Ben Day",
#     "Hayden Garrett",
#     "Kairo McGuire",
#     "April Jenkins",
#     "Declan Cain",
#     "Kendra Rich",
#     "Miller Ellison",
#     "Raina Christensen",
#     "Gregory Browning",
#     "Princess Frazier",
#     "Callum Tanner",
#     "Harmoni Lambert",
#     "Mario Herrera",
#     "Ximena Vega",
#     "Aidan Richmond",
#     "Whitney Cross",
#     "Fabian Decker",
#     "Aleena Michael",
#     "Bronson Brooks",
#     "Autumn Harvey",
#     "Cayden Sandoval",
#     "Elsie Walter",
#     "Lochlan Richardson",
#     "Allison Gentry",
#     "Magnus Hardy",
#     "Jessica Solomon",
#     "Musa Buckley",
#     "Theodora Coleman",
#     "Micah Mathis",
#     "Anne Short",
#     "Hezekiah Logan",
#     "Kora Rollins",
#     "Wes Proctor",
#     "Chandler Portillo",
#     "Wallace Miller",
#     "Isabella O’Donnell",
#     "Lian Dennis",
#     "Maisie Adkins",
#     "Kylo Johnston",
#     "Laila Little",
#     "Lennox Carpenter",
#     "Lilly Lin",
#     "Conor Santiago",
#     "Nyla Khan",
#     "Kendrick Fleming",
#     "Fatima Collier",
#     "Edison Beasley",
#     "Jaylah Simmons",
#     "Harrison Olsen",
#     "Oaklyn Parrish",
#     "Karsyn Yates",
#     "Charley Adams",
#     "Hudson Hardin",
#     "Vada Castaneda",
#     "Collin Barton",
#     "Danna Chan",
#     "Frank House",
#     "Sariah Fleming",
#     "Fernando Baker",
#     "Isla Carpenter",
#     "Jeremy Herring",
#     "Denver Murray",
#     "Ashton Wong",
#     "Adelaide Craig",
#     "Odin Reyes",
#     "Audrey Acosta",
#     "Jensen Middleton",
#     "Madalyn Brandt",
#     "Damir Montgomery",
#     "Evangeline Chen",
#     "Emmanuel Monroe",
#     "Carly Davenport",
#     "Dariel Hammond",
#     "Holly Dorsey",
#     "Enoch Malone",
#     "Skyler Magana",
#     "Rey Torres",
#     "Violet Cruz",
#     "Ryan Hood",
#     "Briana Moses",
#     "Niklaus Walton",
#     "Scarlet Travis",
#     "Willie Curtis",
#     "Alexis Gomez",
#     "Isaiah Douglas",
#     "Aniyah Massey",
#     "Donald Stark",
#     "Kamilah Calhoun",
#     "Gary Larsen",
#     "Xiomara Lynn",
#     "Zechariah Sparks",
#     "Aisha Hail",
#     "Hector Kennedy",
#     "Brianna Caldwell",
#     "Rylan Daniel",
#     "Joy Hoffman",
#     "Steven Wolf",
#     "Jolene Snyder",
#     "Thiago Chandler",
#     "Viviana Mitchell",
#     "Jaxon McClure",
#     "Estella Baldwin",
#     "Jaiden Galindo",
#     "Corinne Hess",
#     "Lawrence Roberson",
#     "Sasha Portillo",
#     "Wallace Carson",
#     "Nalani Jenkins",
#     "Declan Buckley",
#     "Theodora Cole",
#     "Nathaniel Larsen",
#     "Xiomara Herring",
#     "Henrik Rogers",
#     "Madelyn Stephens",
#     "Messiah Turner",
#     "Brooklyn Allison",
#     "Dennis Santiago",
#     "Nyla Patterson",
#     "Amir McKenzie",
#     "Briar Garner",
#     "Sage Horne",
#     "Marlowe Cisneros",
#     "Alden Delarosa",
#     "Iyla Collier",
#     "Edison Peck",
#     "Crystal Sparks",
#     "Drake Meyers",
#     "Leyla Nixon",
#     "Cory Park",
#     "Lia Phillips",
#     "Andrew McCullough",
#     "Hana O’Donnell",
#     "Lian Galindo",
#     "Corinne Juarez",
#     "Joaquin Foster",
#     "Brielle Cline",
#     "Cullen Montoya",
#     "Kamryn Cannon",
#     "Archie French",
#     "Lorelai Lu",
#     "Duncan Choi",
#     "Karla Carrillo",
#     "Wade Tapia",
#     "Michaela Jackson",
#     "Sebastian Edwards",
#     "Ivy Chandler",
#     "Royal Delgado",
#     "Alani Parks",
#     "Gianni Kaur",
#     "Holland Wiley",
#     "Mathew Arroyo",
#     "Kyra Sanders",
#     "Jose Boyer",
#     "Chaya Dejesus",
#     "Rio Frazier",
#     "Octavia Snow",
#     "Houston Fleming",
#     "Fatima Benitez",
#     "Justice Blake",
#     "Amanda Larson",
#     "Rafael Jordan",
#     "Adalynn McLaughlin",
#     "Ibrahim Santos",
#     "Alana Gaines",
#     "Talon Barron",
#     "Anya Pearson",
#     "Gunner English",
#     "Kelly Rosario",
#     "Jedidiah Davenport",
#     "Adrianna Bryan",
#     "Jaxtyn Nguyen",
#     "Nova Peters",
#     "Patrick Farley",
#     "Wrenley Avila",
#     "Jaylen Griffith",
#     "Alicia Macias",
#     "Moshe Summers",
#     "Frankie Barr",
#     "Harley Henson",
#     "Kinslee Chandler",
#     "Royal Esparza",
#     "Ramona Arias",
#     "Alec Wilcox",
#     "Ashlyn Bullock",
#     "Ben Neal",
#     "Talia Pugh",
#     "Judson Sullivan",
#     "Melanie Mullen",
#     "Shepard Blankenship",
#     "Rosalee Barajas",
#     "Brennan Callahan",
#     "Kimber Wong",
#     "Walter Felix",
#     "Paisleigh Mullen",
#     "Shepard Ray",
#     "Ruth McPherson",
#     "Foster Henson",
#     "Kinslee Foley",
#     "Mohammad Salazar",
#     "Freya Nguyen",
#     "Gabriel Hail",
#     "Lainey Norton",
#     "Callen Kent",
#     "Jazmine Hensley",
#     "Layne Walters",
#     "Samara Landry",
#     "Jaxx Merritt",
#     "Kaisley Berry",
#     "Adonis Rich",
#     "Sunny Hurley",
#     "Van Barber",
#     "Cassidy Rasmussen",
#     "Will Watson",
#     "Hailey Moon",
#     "Nova Stephenson",
#     "Khaleesi Shields",
#     "Devon Schwartz",
#     "Lilliana Chang",
#     "Ari Meadows",
#     "Pearl Horn",
#     "Wilson Burke",
#     "Vera Liu",
#     "Pedro Waters",
#     "Bristol Watson",
#     "Greyson Harding",
#     "Aniya Hardy",
#     "Jayceon Rangel",
#     "Gloria Shepherd",
#     "Ronald Nunez",
#     "Mya Parks",
#     "Gianni McMahon",
#     "Belen Blake",
#     "Zyaire Kelly",
#     "Ruby Rivers",
#     "Bear Kane",
#     "Ellianna Hancock",
#     "Rex Douglas",
#     "Aniyah Orozco",
#     "Keanu Calderon",
#     "Serena Barron",
#     "Dustin Acosta",
#     "Kaia Porter",
#     "Rhett Berger",
#     "Laylah Solomon",
#     "Musa Felix",
#     "Paisleigh Wilcox",
#     "Jerry Drake",
#     "Jayleen Carter",
#     "Maverick Tyler",
#     "Helena Felix",
#     "Rodney Lin",
#     "Makenna Briggs",
#     "Case McLaughlin",
#     "Stephanie Harrison",
#     "Gavin Poole",
#     "Bonnie Henson",
#     "Bellamy Mayo",
#     "Aarya Bates",
#     "Ellis Robbins",
#     "Stevie Foster",
#     "Kayden Jacobs",
#     "Camilla Benitez",
#     "Justice King",
#     "Victoria Martinez",
#     "Alexander Russell",
#     "Raelynn Edwards",
#     "Adrian Villanueva",
#     "Monroe Morris",
#     "Christian Strong",
#     "Margo Church",
#     "Terrance Koch",
#     "Milana Crane",
#     "Fox Hayes",
#     "Iris Mendez",
#     "Arthur Sweeney",
#     "Yara Summers",
#     "Darius Griffith",
#     "Alicia Reese",
#     "Alijah Case",
#     "Cleo Leblanc",
#     "Braden Soto",
#     "Brynlee Schroeder",
#     "Izaiah Powers",
#     "Michelle Kirk",
#     "Alessandro Hunt",
#     "Genevieve Johnston",
#     "Felix Joseph",
#     "Gracelynn Pennington",
#     "Bobby Day",
#     "Hayden Reeves",
#     "Clark Hart",
#     "Gemma Park",
#     "Daxton Avery",
#     "Meghan Bartlett",
#     "Kace Curtis",
#     "Alexis Donovan",
#     "Brayan Peck",
#     "Crystal Singh",
#     "Louis Nash",
#     "Novah Ochoa",
#     "Winston Beasley",
#     "Jaylah Henson",
#     "Bellamy Howard",
#     "Sophie Enriquez",
#     "Elisha Spears",
#     "Isabela Vega",
#     "Aidan Lambert",
#     "Nina Chan",
#     "Frank Gibson",
#     "Eden McDonald",
#     "Calvin Price",
#     "Piper O’Connell",
#     "Jovanni Huang",
#     "Francesca Valenzuela",
#     "Jamari Golden",
#     "Giuliana Conway",
#     "Orlando Tran",
#     "Kylie Espinoza",
#     "Dallas Franco",
#     "Charleigh Hamilton",
#     "Jason Farmer",
#     "Madelynn Fleming",
#     "Fernando Shepard",
#     "Noor McFarland",
#     "Dane Pena",
#     "Rachel Browning",
#     "Rohan Gates",
#     "Melina Coleman",
#     "Micah Charles",
#     "Jenna Weber",
#     "Crew Rios",
#     "Brooke Bowman",
#     "Francisco Larson",
#     "Alayna Ray",
#     "Arlo Larsen",
#     "Xiomara Stuart",
#     "Dion Montes",
#     "Roselyn Hughes",
#     "Everett Arias",
#     "Aleah Cook",
#     "Ezekiel Vance",
#     "Maxine Cline",
#     "Cullen Marin",
#     "Celia Daniels",
#     "Xander Curry",
#     "Alison Holloway",
#     "Sutton Bowen",
#     "Dream Lugo",
#     "Santos Torres",
#     "Violet Harrington",
#     "Omari Bell",
#     "Melody Ford",
#     "Luis Barnett",
#     "Harlow Garrett",
#     "Kairo Cross",
#     "Nayeli Logan",
#     "Rocco Wilkinson",
#     "Siena Garcia",
#     "James Bauer",
#     "Haley Sheppard",
#     "Trent Sellers",
#     "Mercy Ashley",
#     "Kylen Randall",
#     "Christina Schmidt",
#     "Zayden Maddox",
#     "Zainab Carpenter",
#     "Jeremy Patterson",
#     "Kaylee Caldwell",
#     "Rylan Kramer",
#     "Hanna Davidson",
#     "Dante Abbott",
#     "Melany Bernal",
#     "Eithan Stuart",
#     "Stormi Frost",
#     "Dario Cummings",
#     "Nylah Madden",
#     "Everest Chandler",
#     "Viviana Salgado",
#     "Trace York",
#     "Milan Clarke",
#     "Stetson Vo",
#     "Artemis Lowe",
#     "Julius Mercado",
#     "Mckinley Delacruz",
#     "Memphis Gillespie",
#     "Alianna Franco",
#     "Gage Monroe",
#     "Carly Petersen",
#     "Samson Weeks",
#     "Karen Harper",
#     "Hayes Cardenas",
#     "Raven Mahoney",
#     "Kamryn Mayo",
#     "Aarya Weber",
#     "Crew Anderson",
#     "Ella Love",
#     "Jeffrey Wallace",
#     "Arianna Cain",
#     "Benson Moon",
#     "Naya Waters",
#     "Maximilian Montoya",
#     "Kamryn May",
#     "Finley Villegas",
#     "Jessie Palacios",
#     "Thaddeus Schultz",
#     "Briella Alvarado",
#     "Andres Carter",
#     "Lucy Davidson",
#     "Dante McCoy",
#     "Mckenzie Stark",
#     "Kristopher Adams",
#     "Stella Ryan",
#     "Timothy Huber",
#     "Raquel Huerta",
#     "Douglas Fisher",
#     "Arya Walters",
#     "Colson Hart",
#     "Gemma Turner",
#     "Joshua Terry",
#     "Wren Stevenson",
#     "Callan Hobbs",
#     "Lacey Hodges",
#     "Alonzo Campbell",
#     "Addison Ashley",
#     "Kylen Blankenship",
#     "Rosalee Kelly",
#     "Cooper Dillon",
#     "Laurel Wise",
#     "Frederick Nguyen",
#     "Nova Romero",
#     "Bryson Meadows",
#     "Pearl Camacho",
#     "Tatum Hines",
#     "Poppy Le",
#     "Damien Sosa",
#     "Cassandra Orozco",
#     "Keanu Wilkinson",
#     "Siena Morales",
#     "Aaron Kirk",
#     "Ellis Henson",
#     "Bellamy Bauer",
#     "Haley May",
#     "Finley Ruiz",
#     "Emery Fox",
#     "Antonio Glass",
#     "Clare Dennis",
#     "Emanuel Dickson",
#     "Emmalynn Reynolds",
#     "Vincent Young",
#     "Zoey Delacruz",
#     "Memphis Morton",
#     "Mallory Dennis",
#     "Emanuel Waller",
#     "Whitley Gonzales",
#     "Brayden Ward",
#     "Ariana Sierra",
#     "Dayton Duran",
#     "Willa Rasmussen",
#     "Will Pineda",
#     "Nola Howe",
#     "Alaric Buck",
#     "Livia Bravo",
#     "Genesis Herring",
#     "Denver Hale",
#     "Ezequiel Cummings",
#     "Nylah Olson",
#     "Malachi Ballard",
#     "Alejandra Day",
#     "Kayson Valentine",
#     "August Barrera",
#     "Makai Boyd",
#     "Georgia Cisneros",
#     "Alden Andrews",
#     "Payton Hudson",
#     "Peter Wells",
#     "Cecilia Daniel",
#     "Grady Griffin",
#     "Charlie Cobb",
#     "Raphael Harmon",
#     "Maren Branch",
#     "Keenan Shepherd",
#     "Katalina Mullins",
#     "Allen Daniel",
#     "Joy Navarro",
#     "Reid Santana",
#     "Myra Miranda",
#     "Rory Stephenson",
#     "Khaleesi Pittman",
#     "Valentino Taylor",
#     "Sofia Kaur",
#     "Augustine Olson",
#     "Isabel Petersen",
#     "Samson Koch",
#     "Milana Lu",
#     "Duncan Espinoza",
#     "Lucille Stein",
#     "Creed Kelly",
#     "Ruby Valdez",
#     "Kyler Singh",
#     "Vivienne Humphrey",
#     "Krew Ryan",
#     "Morgan Newton",
#     "Santino Sosa",
#     "Cassandra Gentry",
#     "Magnus Hurst",
#     "Adalee Melendez",
#     "Nikolas Lambert",
#     "Nina Burton",
#     "Zander Khan",
#     "Mabel Graves",
#     "Cesar Hopkins",
#     "Gabriela Harding",
#     "Brodie Franklin",
#     "Angela McKee",
#     "Bjorn Fletcher",
#     "Anaya Delgado",
#     "Colt Washington",
#     "Valerie Hart",
#     "Joel Klein",
#     "Elianna Bradshaw",
#     "Emory Massey",
#     "Clementine Phillips",
#     "Andrew Hines",
#     "Poppy Patterson",
#     "Amir Peterson",
#     "Caroline Whitehead",
#     "Zayd Richards",
#     "Trinity Figueroa",
#     "Spencer Chambers",
#     "Makayla Clements",
#     "Fisher Campbell",
#     "Addison Gentry",
#     "Magnus Morton",
#     "Mallory Whitney",
#     "Jeffery Collier",
#     "Ivory Benton",
#     "Jamal Stephenson",
#     "Khaleesi Carter",
#     "Maverick Coffey",
#     "Paola Huffman",
#     "Chris Rangel",
#     "Gloria Jenkins",
#     "Declan Williamson",
#     "Catherine Bond",
#     "Roger Patton",
#     "Lorelei Christian",
#     "Ledger Atkinson",
#     "Jazmin Arias",
#     "Alec Santiago",
#     "Nyla Webb",
#     "Lorenzo Hartman",
#     "Kennedi McKinney",
#     "Romeo Cummings",
#     "Nylah Long",
#     "Jace Buchanan",
#     "Maryam Hickman",
#     "Jakobe Hubbard",
#     "Rosie Esparza",
#     "Carl York",
#     "Milan Burgess",
#     "Kolton Zhang",
#     "Sarai Estes",
#     "Hakeem Padilla",
#     "Maggie Kerr",
#     "Louie Beltran",
#     "Kaydence Bailey",
#     "Axel Gates",
#     "Melina Garcia",
#     "James Valenzuela",
#     "Henley Bradshaw",
#     "Emory Baxter",
#     "Lara Vaughan",
#     "Castiel Romero",
#     "Eliza Hebert",
#     "Guillermo Bush",
#     "Everlee Solomon",
#     "Musa Hanna",
#     "Cynthia Benson",
#     "Desmond Harrington",
#     "Legacy Velez",
#     "Kareem Wu",
#     "Liana Houston",
#     "Sylas Hess",
#     "Kaliyah Castaneda",
#     "Collin Sullivan",
#     "Melanie Nava",
#     "Stefan Durham",
#     "Tiffany Williams",
#     "Oliver Johnson",
#     "Emma Curtis",
#     "Muhammad Marsh",
#     "Adelina Cameron",
#     "Rayan Mora",
#     "Jemma Sexton",
#     "Mylo Bravo",
#     "Amoura Barnes",
#     "Damian Corona",
#     "Marianna Patterson",
#     "Amir Parrish",
#     "Tiana Norman",
#     "Aziel Gardner",
#     "Jordyn Everett",
#     "Camilo Wolf",
#     "Jolene Gregory",
#     "Travis Sampson",
#     "Meilani Welch",
#     "Hendrix Person",
#     "Dylan Delarosa",
#     "Osiris Howard",
#     "Sophie Stanton",
#     "Zyair Foley",
#     "Zaylee Burnett",
#     "Davis Crane",
#     "Della Flores",
#     "Lincoln Rubio",
#     "Hadassah Lam",
#     "Bodie Wolfe",
#     "Hallie Cruz",
#     "Ryan Potts",
#     "Ellison Correa",
#     "Zakai McCarthy",
#     "Kira Garrison",
#     "Noe Steele",
#     "Rylie Vu",
#     "Kamdyn Wheeler",
#     "Sydney Finley",
#     "Calum Soto",
#     "Brynlee Castro",
#     "Jasper Higgins",
#     "Leighton Marquez",
#     "Malakai Galindo",
#     "Corinne Melton",
#     "Lennon Beasley",
#     "Jaylah Carpenter",
#     "Jeremy Banks",
#     "Cali Palmer",
#     "Theo Carroll",
#     "Zara Jennings",
#     "Corbin Carr",
#     "Rowan Robinson",
#     "Matthew Huang",
#     "Francesca Felix",
#     "Rodney Barrett",
#     "Kendall Christian",
#     "Ledger Long",
#     "Jade Bernal",
#     "Eithan Erickson",
#     "Sabrina Ochoa",
#     "Winston Malone",
#     "Skyler Kent",
#     "Mekhi Barajas",
#     "Keilani Parks",
#     "Gianni Patrick",
#     "Lyra Jacobs",
#     "Bryan Stephenson",
#     "Khaleesi Howell",
#     "Bradley Sanchez",
#     "Aria Nguyen",
#     "Gabriel Christian",
#     "Anahi Blanchard",
#     "Adler Lawson",
#     "Phoebe McCullough",
#     "Briar Meadows",
#     "Pearl Chase",
#     "Otis Glover",
#     "Alessia Barnes",
#     "Damian Humphrey",
#     "Journi Booth",
#     "Chaim Melton"
# ]
# random_snake_breeds = [
#     "Fishing snake",
#     "Hutton's tree viper",
#     "Collett's snake",
#     "Hagen's pitviper",
#     "Papuan python",
#     "Eastern mud snake",
#     "White-lipped python",
#     "Egg-eater",
#     "Red blood python",
#     "Namaqua dwarf adder",
#     "Inland taipan",
#     "Asian keelback",
#     "Grand Canyon rattlesnake",
#     "Mountain adder",
#     "Dwarf boa",
#     "King cobra",
#     "Red spitting cobra",
#     "Cottonmouth",
#     "Barbour's pit viper",
#     "Western woma python",
#     "Namib dwarf sand adder",
#     "Sand adder",
#     "Malabar rock pitviper",
#     "Aesculapian snake",
#     "Ornate flying snake",
#     "Banded pit viper",
#     "Longnosed worm snake",
#     "King rat snake",
#     "Savu python",
#     "Great Plains rat snake",
#     "Many-horned adder",
#     "Midget faded rattlesnake",
#     "Amazon tree boa",
#     "Bamboo viper",
#     "Hogg Island boa",
#     "Southwestern blackhead snake",
#     "Many-banded tree snake",
#     "Fifty pacer",
#     "Rhombic night adder",
#     "Carpet viper",
#     "Whip Snake",
#     "Long-tailed rattlesnake",
#     "Peninsula tiger snake",
#     "Ball Python",
#     "Bamboo pit viper",
#     "North Philippine cobra",
#     "Chicken snake",
#     "Andaman cat snake",
#     "Forsten's cat snake",
#     "Green tree python",
#     "Texas night snake",
#     "Godman's pit viper",
#     "Horned viper",
#     "Perrotet's shieldtail snake",
#     "Sonoran",
#     "Baja California lyresnake",
#     "Texas lyre snake",
#     "Egyptian cobra",
#     "Green mamba",
#     "Spotted python",
#     "Yellow-bellied sea snake",
#     "Stimson's python",
#     "Indian egg-eater",
#     "Urutu",
#     "Arabian cobra",
#     "Buff striped keelback",
#     "Bimini racer",
#     "Amethystine python",
#     "King cobra",
#     "Scott Shields Barrows Deadliest Snake",
#     "Egyptian asp",
#     "Stiletto snake",
#     "Emerald tree boa",
#     "Twin-Barred tree snake",
#     "Ursini's viper",
#     "Yunnan keelback",
#     "Tiger keelback",
#     "Tropical rattlesnake",
#     "New Guinea carpet python",
#     "Arizona black rattlesnake",
#     "Crowned snake",
#     "Tan racer",
#     "Kaulback's lance-headed pitviper",
#     "Mojave rattlesnake",
#     "Yellow-striped rat snake",
#     "Cape cobra",
#     "Yellow-lined palm viper",
#     "African wolf snake",
#     "Cat-eyed night snake",
#     "Black rat snake",
#     "Death Adder",
#     "River jack",
#     "Coachwhip snake",
#     "Elaps harlequin snake",
#     "Down's tiger snake",
#     "Western tiger snake",
#     "Many-banded krait",
#     "Yellow-lipped sea snake",
#     "Nilgiri keelback",
#     "King snake",
#     "Southern Pacific rattlesnake",
#     "Gaboon viper",
#     "Black-speckled palm-pitviper",
#     "Sumatran tree viper",
#     "Brazilian smooth snake",
#     "Cantil",
#     "Horseshoe pitviper",
#     "Mexican hognose snake",
#     "Twin-spotted rat snake",
#     "Eastern green mamba",
#     "Philippine pitviper",
#     "Mussurana",
#     "Malayan long-glanded coral snake",
#     "Bluntnose viper",
#     "Mollucan python",
#     "Schlegel's viper",
#     "Herald snake",
#     "Flinders python",
#     "Beauty rat snake",
#     "Tasmanian tiger snake",
#     "Puerto Rican boa",
#     "Common ground snake",
#     "Ground snake",
#     "Habu pit viper",
#     "Puff adder",
#     "Leaf viper",
#     "Rosy boa",
#     "South Andaman krait",
#     "Coastal taipan",
#     "Bush viper",
#     "Forest flame snake",
#     "Japanese striped snake",
#     "Beddome's cat snake",
#     "Blunt-headed tree snake",
#     "Eastern water cobra",
#     "Queen snake",
#     "Woma python",
#     "Great Lakes bush viper",
#     "Rough-scaled bush viper",
#     "Burrowing cobra",
#     "Kayaudi dwarf reticulated python",
#     "Okinawan habu",
#     "North eastern king snake",
#     "Dumeril's boa",
#     "Jumping viper",
#     "Hoop snake",
#     "Rufous beaked snake",
#     "Children's python",
#     "Black headed python",
#     "Dauan Island water python",
#     "Chinese tree viper",
#     "Mandalay cobra",
#     "Striped snake",
#     "Mamushi",
#     "Cascabel",
#     "Tokara habu",
#     "Mud adder",
#     "Indian flying snake",
#     "Asian Vine Snake",
#     "Blonde hognose snake",
#     "Texas Coral Snake",
#     "Ribbon snake",
#     "Rattler",
#     "Green palm viper",
#     "Jerdon's pitviper",
#     "Mexican palm-pitviper",
#     "Central American lyre snake",
#     "Tic polonga",
#     "Indonesian water python",
#     "Twig snake",
#     "Harlequin coral snake",
#     "Mexican green rattlesnake",
#     "Rough-scaled tree viper",
#     "Northern white-lipped python",
#     "South American hognose snake",
#     "White-lipped keelback",
#     "Asian cobra",
#     "Indian python",
#     "Modest keelback",
#     "Brazilian coral snake",
#     "Madagascar tree boa",
#     "Sumatran short-tailed python",
#     "Australian copperhead",
#     "Blood python",
#     "False water cobra",
#     "Kaznakov's viper",
#     "Barred wolf snake",
#     "Zebra spitting cobra",
#     "Black-striped keelback",
#     "Eastern hognose snake",
#     "Dubois's sea snake",
#     "Tibetan bamboo pitviper",
#     "Hook Nosed Sea Snake",
#     "Ikaheka snake",
#     "Selayer reticulated python",
#     "Sonoran sidewinder",
#     "Saw-scaled viper",
#     "High Woods coral snake",
#     "Borneo short-tailed python",
#     "De Schauensee's anaconda",
#     "Common worm snake",
#     "False coral snake",
#     "Pallas' viper",
#     "Congo snake",
#     "McMahon's viper",
#     "Sedge viper",
#     "False horned viper",
#     "Burmese krait",
#     "Lyre snake",
#     "Dog-toothed cat snake",
#     "Northern water snake",
#     "Speckled hognose snake",
#     "Tree boa",
#     "Black-banded trinket snake",
#     "Indian cobra",
#     "White-lipped cobra",
#     "Brown spotted pitviper",
#     "Plains hognose snake",
#     "Glossy snake",
#     "Southern hognose snake",
#     "Tree viper",
#     "Mole viper",
#     "Paupan taipan",
#     "Northern black-tailed rattlesnake",
#     "Long-nosed adder",
#     "Sand viper",
#     "Banded Flying Snake",
#     "Western diamondback rattlesnake",
#     "Equatorial spitting cobra",
#     "Boa constrictor",
#     "Rhinoceros viper",
#     "Dwarf pipe snake",
#     "Texas blind snake",
#     "Assam keelback",
#     "Southern black racer",
#     "Sand boa",
#     "Trinket snake",
#     "Patchnose snake",
#     "Western coral snake",
#     "Green tree pit viper",
#     "Anaconda",
#     "Mandarin rat snake",
#     "Dusky pigmy rattlesnake",
#     "Raddysnake",
#     "Southwestern black spitting cobra",
#     "Spitting cobra",
#     "Cyclades blunt-nosed viper",
#     "Kanburian pit viper",
#     "Amazonian palm viper",
#     "Peringuey's adder",
#     "Honduran palm viper",
#     "Night snake",
#     "Stoke's sea snake",
#     "Colorado desert sidewinder",
#     "Red-tailed bamboo pitviper",
#     "Desert kingsnake",
#     "Rainbow boa",
#     "Jamaican boa",
#     "Black-tailed horned pit viper",
#     "Gold's tree cobra",
#     "Eastern diamondback rattlesnake",
#     "Rough-scaled python",
#     "Tiger pit viper",
#     "Common tiger snake",
#     "Buttermilk racer",
#     "Arizona coral snake",
#     "Jararacussu",
#     "Tawny cat snake",
#     "Great Basin rattlesnake",
#     "Palestinian viper",
#     "Blue krait",
#     "Smooth green snake",
#     "Beddome's coral snake",
#     "Worm snake",
#     "Brown water python",
#     "Northwestern carpet python",
#     "Red-headed krait",
#     "Chinese cobra",
#     "Nicobar bamboo pitviper",
#     "Brazilian mud Viper",
#     "Guatemalan tree viper",
#     "Indochinese spitting cobra",
#     "Lance-headed rattlesnake",
#     "Sinai desert cobra",
#     "Congo water cobra",
#     "Shield-nosed cobra",
#     "Coronado Island rattlesnake",
#     "Western hognose snake",
#     "Hill keelback",
#     "Eyelash palm-pitviper",
#     "Rough green snake",
#     "Sri Lankan pit viper",
#     "Burmese keelback",
#     "Cuban boa",
#     "Cobra de capello",
#     "Orange-collared keelback",
#     "Brown snake",
#     "Machete savane",
#     "Coastal carpet python",
#     "Twin-spotted rattlesnake",
#     "Caspian cobra",
#     "Twin Headed King Snake",
#     "Fer-de-lance",
#     "Khasi Hills keelback",
#     "Hairy bush viper",
#     "Nightingale adder",
#     "Durango rock rattlesnake",
#     "Royal python",
#     "Eastern tiger snake",
#     "Australian scrub python",
#     "Hundred pacer",
#     "Jamaican Tree Snake",
#     "Forest cobra",
#     "Thai cobra",
#     "Boomslang",
#     "Black krait",
#     "Large-eyed pitviper",
#     "Bocourt's water snake",
#     "Bushmaster",
#     "Siamese palm viper",
#     "Pine snake",
#     "Common cobra",
#     "Ceylon krait",
#     "American copperhead",
#     "Mangrove snake",
#     "Lesser black krait",
#     "Andrea's keelback",
#     "Ball python",
#     "Cascavel",
#     "Eastern yellowbelly sad racer",
#     "Burrowing viper",
#     "Wagler's pit viper",
#     "West African brown spitting cobra",
#     "Eastern hognose snake",
#     "Banded water cobra",
#     "Desert woma python",
#     "Prairie kingsnake",
#     "King Island tiger snake",
#     "Brown tree snake",
#     "Uracoan rattlesnake",
#     "Himalayan keelback",
#     "Black mamba",
#     "Persian rat snake",
#     "Tri-color hognose snake",
#     "Wetar Island python",
#     "Tentacled snake",
#     "Eye-lash viper",
#     "Long-nosed tree snake",
#     "Sunbeam snake",
#     "Lancehead",
#     "Spectacle cobra",
#     "Timber rattlesnake",
#     "Leaf-nosed viper",
#     "Sharp-nosed viper",
#     "Yellow cobra",
#     "Schultze's pitviper",
#     "Tigre snake",
#     "Side-striped palm-pitviper",
#     "Olive python",
#     "Wolf snake",
#     "Spiny bush viper",
#     "Red diamond rattlesnake",
#     "Rinkhals cobra",
#     "Hognosed viper",
#     "Red-tailed pipe snake",
#     "Speckle-bellied keelback",
#     "Pelagic sea snake",
#     "Paradise flying snake",
#     "Mexican west coast rattlesnake",
#     "Pipe snake",
#     "Parrot snake",
#     "Blind snake",
#     "Brown white-lipped python",
#     "Zebra snake",
#     "Sikkim keelback",
#     "Javan spitting cobra",
#     "Tanimbar python",
#     "Baird's rat snake",
#     "Madagascar ground boa",
#     "Bull snake",
#     "Black-necked cobra",
#     "Red-tailed boa",
#     "Wall's keelback",
#     "Crossed viper",
#     "Diamond python",
#     "Brahminy blind snake",
#     "Rattlesnake",
#     "Grey lora",
#     "Cuban wood snake",
#     "Scarlet kingsnake",
#     "False cobra",
#     "Nose-horned viper",
#     "Dwarf sand adder",
#     "Rungwe tree viper",
#     "Arafura file snake",
#     "Cat-eyed snake",
#     "Angolan python",
#     "Southern Philippine cobra",
#     "Tiger rattlesnake",
#     "Hopi rattlesnake",
#     "Mud snake",
#     "Elegant pitviper",
#     "Common keelback",
#     "Eastern coral snake",
#     "Horned desert viper",
#     "Halmahera python",
#     "Western blind snake",
#     "Flat-nosed pitviper",
#     "Nitsche's tree viper",
#     "Gopher snake",
#     "Montpellier snake",
#     "Levant viper",
#     "Small-eyed snake",
#     "Eastern lyre snake",
#     "Black tree cobra",
#     "Mojave desert sidewinder",
#     "Mangrove pit viper",
#     "South eastern corn snake",
#     "Many-spotted cat snake",
#     "Large-scaled tree viper",
#     "Wirot's pit viper",
#     "Long-nosed whip snake",
#     "Checkered garter snake",
#     "Gold-ringed cat snake",
#     "Long-nosed viper",
#     "European smooth snake",
#     "Desert death adder",
#     "Banded krait",
#     "Jungle carpet python",
#     "Burmese python",
#     "Central ranges taipan",
#     "Pope's tree viper",
#     "Southwestern carpet python",
#     "Chihuahuan ridge-nosed rattlesnake",
#     "West Indian racer",
#     "Manchurian Black Water Snake",
#     "Massasauga rattlesnake",
#     "Guatemalan palm viper",
#     "Tancitaran dusky rattlesnake",
#     "African rock python",
#     "Red-necked keelback",
#     "Harlequin snake",
#     "Brongersma's pitviper",
#     "Grey-banded kingsnake",
#     "White-lipped tree viper",
#     "Snouted cobra",
#     "Blanding's tree snake",
#     "Nichell snake",
#     "Philippine cobra",
#     "Russell's viper",
#     "Eastern brown snake",
#     "Green anaconda",
#     "Olive sea snake",
#     "Eastern racer",
#     "Krefft's tiger snake",
#     "Oenpelli python",
#     "Northeastern hill krait",
#     "King brown",
#     "Aruba rattlesnake",
#     "Kham Plateau pitviper",
#     "Large shield snake",
#     "Black snake",
#     "Lora",
#     "Bredl's python",
#     "Whip snake",
#     "Nose-horned viper",
#     "Mexican black kingsnake",
#     "Yarara",
#     "African twig snake",
#     "Black-headed snake",
#     "California kingsnake",
#     "Golden tree snake",
#     "Rubber boa",
#     "Annulated sea snake",
#     "Green cat-eyed snake",
#     "Monocled cobra",
#     "Indian krait",
#     "Eyelash pit viper",
#     "Rinkhals",
#     "Fierce snake",
#     "Wynaad keelback",
#     "Indian tree viper",
#     "Garter snake",
#     "Yellow anaconda",
#     "Water snake",
#     "Nitsche's bush viper",
#     "Western hog-nosed viper",
#     "Mozambique spitting cobra",
#     "Abaco Island boa",
#     "Cape gopher snake",
#     "Krait",
#     "Moluccan flying snake",
#     "Fan-Si-Pan horned pitviper",
#     "Fea's viper",
#     "Beaked sea snake",
#     "Bird snake",
#     "Grass snake",
#     "Boiga",
#     "Water adder",
#     "Nicobar cat snake",
#     "Horned adder",
#     "Checkered keelback",
#     "Speckled kingsnake",
#     "Asp viper",
#     "Bolivian anaconda",
#     "Ringed hognose snake",
#     "Sri Lanka cat snake",
#     "Wutu",
#     "Milk snake",
#     "Jan's hognose snake",
#     "Eastern fox snake",
#     "Cape coral snake",
#     "Red-bellied black snake",
#     "European asp",
#     "Gray cat snake",
#     "Bismarck ringed python",
#     "Southern white-lipped python",
#     "Oaxacan small-headed rattlesnake",
#     "Malayan pit viper",
#     "Three-lined ground snake",
#     "Boelen python",
#     "African puff adder",
#     "Timor python",
#     "African beaked snake",
#     "Pygmy python",
#     "Mexican racer",
#     "Titanboa snake",
#     "Giant Malagasy hognose snake",
#     "Shield-tailed snake",
#     "Monoculate cobra",
#     "Reticulated python",
#     "Black-necked spitting cobra",
#     "Mangshan pitviper",
#     "Japanese rat snake",
#     "San Francisco garter snake",
#     "Snorkel viper",
#     "Calabar python",
#     "Temple viper",
#     "Moccasin snake",
#     "Southern Indonesian spitting cobra",
#     "Dwarf beaked snake",
#     "Western ground snake",
#     "Himehabu",
#     "Andaman cobra",
#     "Cantor's pitviper",
#     "Western mud snake",
#     "Inland carpet python",
#     "Green snake",
#     "Western green mamba",
#     "Japanese forest rat snake",
#     "Asian pipe snake",
#     "Dusty hognose snake",
#     "Portuguese viper",
#     "Eyelash viper",
#     "Common adder",
#     "American Vine Snake",
#     "Banded cat-eyed snake",
#     "River jack",
#     "Hardwicke's sea snake",
#     "Canebrake",
#     "Green rat snake",
#     "Sakishima habu",
#     "Mexican vine snake",
#     "Texas garter snake",
#     "Motuo bamboo pitviper",
#     "Centralian carpet python",
#     "Mexican parrot snake",
#     "Sind krait",
#     "Malayan krait",
#     "Vine snake",
#     "Bornean pitviper",
#     "Water moccasin",
#     "Nubian spitting cobra",
#     "Undulated pit viper",
#     "Wart snake",
#     "Malcolm's tree viper",
#     "Common garter snake",
#     "Storm water cobra",
#     "Macklot's python",
#     "Western carpet python",
#     "Northern tree snake",
#     "Nicobar Island keelback",
#     "Temple pit viper",
#     "Common lancehead",
#     "Stejneger's bamboo pitviper",
#     "Yellow-banded sea snake",
#     "Eastern coral snake",
#     "Narrowhead Garter Snake",
#     "Southwestern speckled rattlesnake",
#     "Chappell Island tiger snake"
# ]
# random_colours = [
#     "Green",
#     "Brown",
#     "Black",
#     "Yellow",
#     "Red",
#     "Orange"
# ]

# snake_collection = []

# for x in range(1000):
#     snake_name = random_names[random.randint(0, (len(random_names) - 1))]
#     snake_colour = random_colours[random.randint(0, (len(random_colours) - 1))]
#     snake_breed = random_snake_breeds[random.randint(0, (len(random_snake_breeds) - 1))]
#     snake_age = random.randint(1, 12)

#     new_snake = Snake(snake_name, snake_colour, snake_breed, snake_age)
#     untold.add_table_row("snakes", new_snake.__dict__)

# untold.create_db_table("snakes", untold.generate_db_table_schema(Snake), "name")

# untold.get_table_row("snakes", "*")

# print(tester.__dict__)
# print(tester.__annotations__)
# print(untold.generate_db_table_schema(UserObj))

# untold.login("admin1", "admin")
# untold.add_table_row("snakes", {
#     "Name": "Mrs. Hiss",
#     "Colour": "Brown",
#     "Breed": "Carpet Python"
# })
# untold.update_table_row("snakes", "*")
# untold.delete_table_row("snakes", "*")
# untold.get_table_row("snakes", "*")

# u.create_db_table("snakes", [
#     {
#         "ColumnName": "Snake_ID",
#         "ColumnType": "string",
#         "Nullable": False
#     },
#     {
#         "ColumnName": "Name",
#         "ColumnType": "string",
#         "Nullable": False
#     },
#     {
#         "ColumnName": "Colour",
#         "ColumnType": "string",
#         "Nullable": False
#     },
#     {
#         "ColumnName": "Breed",
#         "ColumnType": "string",
#         "Nullable": False
#     }
# ], "Snake_ID")
