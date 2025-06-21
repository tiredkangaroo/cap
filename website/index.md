# cap
cap is a configurable proxy server that allows you to capture, modify, and inspect HTTP and HTTPS traffic.

<!-- # Screenshots
![1](https://raw.githubusercontent.com/tiredkangaroo/bigproxy/refs/heads/main/screenshots/1.png)
A request being displayed in the UI with two filters selected.

![2](https://raw.githubusercontent.com/tiredkangaroo/bigproxy/refs/heads/main/screenshots/2.png)
A request being modified in the UI (edit mode) while also asking for approval (Wait for Approval).

![3](https://raw.githubusercontent.com/tiredkangaroo/bigproxy/refs/heads/main/screenshots/3.png)
A request's response being shown with the status code, headers, body, bytes transferred, and timing.

![4](https://raw.githubusercontent.com/tiredkangaroo/bigproxy/refs/heads/main/screenshots/4.png)
Settings menu. -->


<!-- # Major Features
- Capture HTTP and HTTPS traffic (with MITM support)
- Modify requests (including the destination Host, method, headers, body, etc.)
- See the timeline for the request from start to finish
- Configure the behavior of the proxy server (MITM, Body Dumping, etc.)
- Configure which fields to display in the UI
- Persistent storage of requests
- Filter requests by Client Application, Client IP, Host, with more filters coming soon -->

# Installation
well currently the built version of the app doesn't work, so we're going to have to run it from source.

1. #### Ensure you have Go (1.24+) installed on your system. See the [Go installation guide](https://go.dev/doc/install) for instructions.

2. #### Clone the repository:
   ```bash
   git clone https://github.com/tiredkangaroo/bigproxy.git
   ```

3. #### Navigate to the project directory (change the working directory):
   ```bash
   cd bigproxy
    ```
4. #### Install the dependencies: Make sure you have Node.js and npm installed for the UI dependencies. You can download them from [Node.js official website](https://nodejs.org/).


    Go dependencies:
     ```bash
   go mod tidy
    ```
    UI dependencies:
    ```bash
   npm i --prefix manager
   ```
5. #### Optional: If you want to utilize the Man-in-the-Middle (MITM) functionality, you will need to generate a CA certificate.
   ```bash
   go run make.go gen-ca
   ```
    This will create a CA certificate in the `certs` directory. You can then install this certificate in your system or browser to enable HTTPS traffic interception. You will need to trust this certificate in the system or browser to avoid security warnings when intercepting HTTPS traffic.


6. #### Run the project.
    ```bash
    go run make.go run
    ```
    Note: If you want to run the project in debug mode, you can use:
    ```bash
    go run make.go debug
    ```
7. #### The proxy is now running on `http://localhost:8000`. Configure your browser, application, or system to use this proxy server for HTTP and HTTPS traffic.

# Usage
Once the proxy server is running, you can start capturing and modifying HTTP and HTTPS traffic. Here are some instructions for some basic features:

1. **Configure your browser or application** to use the proxy server:
   - Set the HTTP and HTTPS proxy to `localhost:8000`.

2. **Modify Proxy Behavior**:
    - Press the settings button in the bottom right corner of the UI to configure the proxy behavior, such as enabling MITM, body dumping, etc.

3. **Filter Requests**:
    - Use the filter options in top left of the UI to filter requests by Client Application, Client IP, Host, etc. You can select multiple filters to narrow down the displayed requests.

4. **Modify Requests**:
    1. Ensure the proxy is running with the "Wait for Approval" option enabled.
    2. When a request is captured, it will be displayed in the UI.
    3. Click on the request and press the "Edit" button to modify the request.
    4. Make the desired changes to the request. See below on how to make certain modifications.
    5. Press the save button to apply the changes.
    6. Approve the request by pressing the "Approve" button. The request will then be sent to the server with the modifications.

5. **Making non-self explanatory modifications**
    - **Change the Method**: You can change the HTTP method (GET, POST, etc.) of the request by pressing the method and alternating through the available methods.

    - **Modify Body**: You can modify the body of the request. Press the show body button to view the body, if it is not already shown, and then edit it as needed. You can also change the content type of the body by setting the `Content-Type` header in the request editor. COMING SOON: You will be provided an editor to modify the body in a manner consistent with the content type. e.g JSON editor for JSON bodies, etc.

    - **Modify Perform in HTTP/HTTPS mode**: You can change the request to be performed in HTTP or HTTPS mode by pressing the lock next to the method in the Request card. This will toggle the request between HTTP and HTTPS modes. Note: changing performance to HTTP mode will still not allow strong information to be available to the ui if the request was originally made in HTTPS mode WITH tunneling. The request will continue to be performed in tunnel, but the HTTP/HTTPS mode will be changed to reflect the changed request.

6. **Freeze Requests**:
    - You can freeze the UI in order to allow you to work with the UI without the requests being updated in real-time. This is useful when you want to work with the UI without being interrupted by new requests coming in. Updates that occur are kept but not shown until the unfreeze button is pressed. To freeze the UI, press the snowflake button in the bottom right corner of the UI. You can unfreeze the UI by pressing the same button again.
# Glossory
- **MITM** - Man in the Middle. A technique used to intercept and modify communication between two parties without their knowledge. Using man-in-the-middle, the proxy is able to intercept the full traffic between the client and the server, and provide **strong** information for HTTPS requests as well.
- **Weak information** - The lower level of information the proxy is able to provide. This includes the host, client IP, client application, and the bytes transferred. All requests at the very least provide weak information.
- **Strong information** - The highest level of information the proxy is able to provide. This includes all weak information but also includes the method, headers, query, path, body, etc. as well as the full response including the status code, headers, body, etc. All HTTP requests are capable of providing strong information, while HTTPS requests are only capable of providing strong information if the proxy is running with MITM enabled.
- **Tunneling** - The technique used to connect the client with the server in HTTPS mode without MITM. This is done by establishing a tunnel that copies the encrypted bytes from the client to the server and vice versa. The proxy is not able to see the full request or response in this case, and can only provide weak information. This is the default mode for HTTPS requests.
