# TECHNICAL TEST
## supporteng API analysis

A review of the API code was conducted, with tests in Postman for the detection and logging of potential errors. Evaluations were performed for each endpoint, analyzing the requested parameters (both in the path and the query), the obtained responses, and the implementation in the application code.


The analysis of these tests is broken down into three components:

- General recommendations.
- Endpoint verification.
- Conclusions.


**Table of Contents**
- [TECHNICAL TEST](#technical-test)
  - [supporteng API analysis](#supporteng-api-analysis)
    - [General recommendations](#general-recommendations)
    - [Endpoints Verification](#endpoints-verification)
      - [User creation](#user-creation)
      - [Login](#login)
      - [Get user](#get-user)
      - [Transfer money](#transfer-money)
      - [User report](#user-report)
      - [Get all users](#get-all-users)
    - [Conclusions](#conclusions)

##### GENERAL RECOMMENDATIONS
The following general recommendations are provided to consistently improve the API:

- Apply the appropriate HTTP method based on the type of request and necessary validation. For instance, in the money transfer function, a GET method is currently used, which is meant for reading or querying resources. Instead, an update resource method, such as PATCH or PUT, should be utilized to modify the "money" field in both the source and destination accounts.

- Improve the authentication system; implementing a system of temporary tokens can enhance account security.

- Address the significant deficiency in terms of roles and authorizations within the API. Currently, there is a risk of unintended privilege escalation, allowing any user to make modifications or access other users' accounts. This issue arises due to the absence of access level implementation, resulting in standard users having unrestricted access to and modification rights for information associated with accounts other than their own.

- Store passwords using hash and salt functions, encrypting them for enhanced application security.

- In certain situations, incorrect error messages have been observed. For example, in the "get user" endpoint, an "invalid token" error is generated, returned with a code of 400, although it should be 401 as it is an authentication error.


##### ENDPOINTS VERIFICATION

An analysis is carried out for each endpoint, conducting tests in Postman and reviewing the code with the aim of identifying potential flaws in the application. Once the errors are identified, relevant solution suggestions are provided for each detected error. While these solutions are presented in a theoretical manner, the intention is to offer a guide that enables the developer to address the issue appropriately. Additionally, the link to the application repository with the code modifications to address the detected errors is also provided

###### User creation
***Validation error***: No validations are performed on repeated usernames.
IMAGEN 1
![](https://github.com/Miguelopezlo/supporteng/blob/master/ss/imagen1.PNG)
curl --location --request GET 'http://localhost:4000/users' \
--data '{"Username": "usuario-prueba", "Password": "1234"}'
***Proposed solution***: Implement a validation on the backend before user creation to check if the chosen 'username' already exists. In such a case, notify the user to choose a different 'username'.

###### Login
***Implementation error***: The password is transmitted through the path, which is not recommended as it is an insecure practice. The password may be exposed in various locations, such as the browser history, server logs, and shared links. This poses a significant security risk, as URLs are often stored and may be accessible to others.
IMAGEN 2
![](https://github.com/Miguelopezlo/supporteng/blob/master/ss/imagen2.PNG)
curl --location –request GET 'http://localhost:4000/users/miguel/login?password=1234' \
***Possible solution***: Instead, it is advisable to send passwords through secure and standard methods, such as the body of a POST request with HTTPS (TLS/SSL) to protect information during transmission. It is also recommended to ensure that communication between the client and the server is encrypted using the HTTPS protocol. This will safeguard confidential information, such as passwords, during the transfer between the browser and the server.

###### Get user
***Authentication error***: User information can be accessed solely with the 'username,' meaning there is no implementation of authentication or security. The response includes the password of the queried user, allowing the retrieval of access credentials for any user.
IMAGEN 3
![](https://github.com/Miguelopezlo/supporteng/blob/master/ss/imagen3.PNG)
curl --location --request GET 'http://localhost:4000/users/carlos' \
***Possible solution***: Implement an authentication system that verifies whether the user has permission to access the requested resources through the access token. This should include checking that the user attempting to access another user's information has the necessary permissions.

###### Transfer money
***Validation error***: There is no validation of the value to be sent, allowing transactions of negative values to withdraw money from external accounts solely with the access token. In other words, no authentication of the destination account is required to withdraw money from that account.
IMAGEN 4
![](https://github.com/Miguelopezlo/supporteng/blob/master/ss/imagen4.PNG)
curl --location --request GET http://localhost:4000/users/miguel/transfer?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6Im1pZ3VlbCJ9.KUXPAEnkIgKksTZFFvBSxoqDpq-i9lUCYdIC_BaYkhg&to=nombre-de-usuario&amount=-5010' \
***Possible solution***: It is necessary to implement validation for the information received from the client. In this case, ensure that the transfer amount is a positive value greater than 0, and that the source account has in the 'money' field an amount equal to or greater than the one being transferred.

***Authorization error***: With an access token, changing the username in the path allows me to move money from another user's account. For this example, the access token of the user 'carlos' is used to transfer money from the account of the user 'username'.
IMAGEN 5
![](https://github.com/Miguelopezlo/supporteng/blob/master/ss/imagen5.PNG)
curl --location --request GET 'http://localhost:4000/users/miguel/transfer?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6Im1pZ3VlbCJ9.KUXPAEnkIgKksTZFFvBSxoqDpq-i9lUCYdIC_BaYkhg&to=nombre-de-usuario&amount=-5010' \
***Possible solution***: Implement an authentication system that validates that the token used for the transaction belongs to the 'username' of the source account.

***Implementation error***: The transfer operation is not performed atomically, meaning it is done in parts. First, the money is added to the destination account, and then it is subtracted from the source account.
CODE HERE
***Possible solution***: This operation should be performed simultaneously to ensure the atomicity of the transaction and avoid potential errors during execution.

***Validation error***: There is no validation to ensure that the destination account is different from the source account. When an operation of this type is performed, money is deducted from the account, and it is lost.
IMAGEN 6
![](https://github.com/Miguelopezlo/supporteng/blob/master/ss/imagen6.PNG)
curl --location --request GET 'http://localhost:4000/users/john/transfer?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6Im1pZ3VlbCJ9.KUXPAEnkIgKksTZFFvBSxoqDpq-i9lUCYdIC_BaYkhg&amount=5010.2&to=john' \
***Possible solution***: Implement the necessary validation for the destination field, ensuring that it must be different from the source account.

###### User report
***Authentication error***: The endpoint for generating user reports does not function as depicted in the documentation. Instead, the report requires only the 'id' of the user and the desired format. In other words, a mechanism to secure user report data has not been implemented.
IMAGEN 7
![](https://github.com/Miguelopezlo/supporteng/blob/master/ss/imagen7.PNG)
curl --location 'http://localhost:4000/users/01HR2QJWE3EFK0P6F9YT28PF7B.html' 
***Possible solution***: Implement authentication using the user token.

###### Get all users
***Documentation error***: The endpoint for retrieving all users is not documented, although it is enabled and available. This could lead to potential issues as someone may discover this endpoint and use it incorrectly or maliciously. This endpoint lacks authentication and authorization validations, allowing anyone to obtain a list of all users with their respective IDs.
IMAGEN8
![](https://github.com/Miguelopezlo/supporteng/blob/master/ss/imagen8.PNG)
curl --location 'http://localhost:4000/users'
Possible solution: Document the endpoint for proper usage or deactivate it to prevent unwanted access. Additionally, it is necessary to implement an authorization measure to prevent exposing the information of all users to unauthorized users.

##### CONCLUSIONS:
An assessment of the API was conducted based on the following criteria:
- Functionality: Examining whether the endpoint fulfills its intended purpose.
  - Create user: Fulfills
  - Login: fulfills
  - Get user: fulfills
  - Transfer money: Does not comply
  - User report: Does not comply
  
- Security: Verifying if the endpoint implements security measures such as secure connections, data non-exposure, SQL injection risk mitigation, authentication measures, authorization levels, among others.
  - Create user: Does not comply
  - Login: Does not comply
  - Get user: Does not comply
  - Transfer money: Does not comply
  - User report: Does not comply
    
- Usability: Analyzing the API code to check for the implementation of best practices that facilitate easy and understandable usage for other developers or users. This includes detailed error handling, consistent naming conventions, clear and concise documentation, among other aspects."
  - Create user: Does not comply
  - Login: Does not comply
  - Get user: Does not comply
  - Transfer money: Does not comply
  - User report: Does not comply



