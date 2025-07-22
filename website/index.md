# About

cap is a configurable proxy server that allows you to capture, modify, and inspect HTTP and HTTPS traffic.

# Features

- fast and lightweight proxy server (HTTP and HTTPS)
- capture proxy traffic with the client
- retrieve and modify requests (including the destination, method, headers, body, etc.)\*\*
- filtering requests
- timeline for the request
- more im a bit lazy rn to write all of it

# Installation

well currently the built version of the app doesn't work, so we're going to have to run it from source.

1. #### Clone the repository:

   ```bash
   git clone https://github.com/tiredkangaroo/cap.git
   ```

2. #### Navigate to the project directory (change the working directory):
   ```bash
   cd cap
   ```
3. #### Setup.
   Ensure you have [Go](https://go.dev/doc/install) (1.24+), [Node.js](https://nodejs.org/) installed on your system, and [npm](https://www.npmjs.com/get-npm) (Node Package Manager) installed.
   ```bash
   go run . setup
   ```
4. #### Run the project.

   ```bash
   go run . run
   ```

   If you want to run the project in debug mode, you can use:

   ```bash
   go run . debug
   ```

   **Build the app (MacOS only)**: If you want to build the app for MacOS, you can use the following command:

   ```bash
   go run . app
   ```

   The built app will be located in the working directory as `cap.app`.

5. #### The proxy is now running on `http://localhost:8000`. Configure your browser, application, or system to use this proxy server for HTTP and HTTPS traffic.

Here's a guide for MacOS.

1. Open System Settings and press Network.

   ![8](https://raw.githubusercontent.com/tiredkangaroo/cap/refs/heads/main/screenshots/8.png)

2. Press the network you are using (e.g., Wi-Fi).

   ![9](https://raw.githubusercontent.com/tiredkangaroo/cap/refs/heads/main/screenshots/9.png)

3. Press the "Details" button next to the network name.

   ![10](https://raw.githubusercontent.com/tiredkangaroo/cap/refs/heads/main/screenshots/10.png)

4. Press the "Proxies" tab.

   ![11](https://raw.githubusercontent.com/tiredkangaroo/cap/refs/heads/main/screenshots/11.png)

5. Enable the Secure Web Proxy (HTTPS) option.

   ![12](https://raw.githubusercontent.com/tiredkangaroo/cap/refs/heads/main/screenshots/12.png)

6. Fill in the host as `localhost` and the port to `8000`.

   ![13](https://raw.githubusercontent.com/tiredkangaroo/cap/refs/heads/main/screenshots/13.png)

7. Do the same for the Web Proxy (HTTP) option.

   ![14](https://raw.githubusercontent.com/tiredkangaroo/cap/refs/heads/main/screenshots/14.png)

8. Press the "OK" button to save the changes.

   ![15](https://raw.githubusercontent.com/tiredkangaroo/cap/refs/heads/main/screenshots/15.png)

9. You may now close the System Settings window. Your proxy is now configured and running.

# Usage

Once the proxy server is running, you can start capturing and modifying HTTP and HTTPS traffic. Here are some instructions for some basic features:

1. **Configure your browser or application** to use the proxy server:

   - Set the HTTP and HTTPS proxy to `localhost:8000`.

2. **Settings**:
   - Open the settings menu by clicking on the gear icon in the top right corner of the UI.
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

# Issues

If you encounter any issue or have suggestions for improvements, please open an issue on the [GitHub repository](https://tiredkangaroo.github.io/cap/issues).

You may also contact me via [email](mailto:ajinest6@gmail.com).

# Glossory

- **MITM** - Man in the Middle. A technique used to intercept and modify communication between two parties without their knowledge. Using man-in-the-middle, the proxy is able to intercept the full traffic between the client and the server, and provide **strong** information for HTTPS requests as well.
- **Strong information** - The highest level of information the proxy is able to provide. This includes all weak information but also includes the method, headers, query, path, body, etc. as well as the full response including the status code, headers, body, etc. All HTTP requests are capable of providing strong information, while HTTPS requests are only capable of providing strong information if the proxy is running with MITM enabled.
- **Tunneling** - The technique used to connect the client with the server in HTTPS mode without MITM. This is done by establishing a tunnel that copies the encrypted bytes from the client to the server and vice versa. The proxy is not able to see the full request or response in this case, and can only provide weak information. This is the default mode for HTTPS requests.
- **Weak information** - The lower level of information the proxy is able to provide. This includes the host, client IP, client application, and the bytes transferred. All requests at the very least provide weak information.

\*\* Requires the request to be an HTTP or an HTTPS request with MITM enabled.
