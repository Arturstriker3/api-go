const net = require("net");
const readline = require("readline");
const fs = require("fs").promises;
const path = require("path");

// Load configuration
const config = require("./test-config.json");

// Create interactive CLI
const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
});

// Load email template
async function loadTemplate() {
  const templatePath = path.join(__dirname, "email-template.json");
  const template = await fs.readFile(templatePath, "utf8");
  return JSON.parse(template);
}

// Replace template variables
function processTemplate(template, testNumber) {
  return {
    subject: template.subject,
    body: template.body
      .replace("{{timestamp}}", new Date().toLocaleString())
      .replace("{{testNumber}}", testNumber.toString()),
  };
}

// Sleep function for delays
function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

// Function to send a single email and wait for response
function sendEmailAndWaitResponse(client, email) {
  return new Promise((resolve) => {
    const responseHandler = (data) => {
      try {
        const response = JSON.parse(data.toString());
        client.removeListener("data", responseHandler);
        resolve(response);
      } catch (err) {
        console.error("Error parsing response:", err);
        resolve(null);
      }
    };

    client.on("data", responseHandler);
    client.write(JSON.stringify(email) + "\n");
  });
}

// Function to send a single email
async function sendSingleEmail(template) {
  return new Promise((resolve) => {
    const client = new net.Socket();

    client.connect(config.connection.port, config.connection.host, async () => {
      console.log("\nConnected to GoMailer");

      // Send authentication and wait for response
      const authResponse = await sendEmailAndWaitResponse(client, {
        secret: config.connection.authSecret,
      });

      if (!authResponse || authResponse.error) {
        console.error(
          "\n❌ Authentication failed:",
          authResponse?.error || "No response"
        );
        client.destroy();
        resolve();
        return;
      }

      console.log("✓ Authentication successful");

      // Process template and send email
      const processedTemplate = processTemplate(template, 1);
      const email = {
        to: [config.email.to],
        subject: processedTemplate.subject,
        body: processedTemplate.body,
      };

      const response = await sendEmailAndWaitResponse(client, email);

      if (response && !response.error) {
        console.log("\n✅ Email sent successfully!");
        console.log(`📧 Recipient: ${email.to}`);
        console.log(`📑 Subject: ${email.subject}`);
      } else {
        console.log("\n❌ Failed to send email");
        console.log("Error:", response?.error || "No response");
      }

      client.destroy();
      resolve();
    });

    client.on("error", (err) => {
      console.error("\n❌ Connection error:", err);
      resolve();
    });
  });
}

// Function to send multiple test emails
async function sendTestEmails(template, count) {
  return new Promise((resolve) => {
    const client = new net.Socket();
    let emailsSent = 0;
    let emailErrors = 0;

    client.connect(config.connection.port, config.connection.host, async () => {
      console.log("\nConnected to GoMailer");
      console.log(`Number of emails to send: ${count}`);

      // Send authentication and wait for response
      const authResponse = await sendEmailAndWaitResponse(client, {
        secret: config.connection.authSecret,
      });

      if (!authResponse || authResponse.error) {
        console.error(
          "\n❌ Authentication failed:",
          authResponse?.error || "No response"
        );
        client.destroy();
        resolve();
        return;
      }

      console.log("✓ Authentication successful");
      console.log("\nStarting email batch...\n");

      // Send emails one by one, waiting for response and adding delay
      for (let i = 1; i <= count; i++) {
        const processedTemplate = processTemplate(template, i);
        const email = {
          to: [config.email.to],
          subject: processedTemplate.subject,
          body: processedTemplate.body,
        };

        const response = await sendEmailAndWaitResponse(client, email);
        if (response && !response.error) {
          emailsSent++;
          console.log(`✓ Email ${i}/${count} sent successfully`);
        } else {
          emailErrors++;
          console.error(
            `❌ Failed to send email ${i}/${count}:`,
            response?.error || "No response"
          );
        }

        // Add delay between sends (500ms)
        if (i < count) {
          await sleep(500);
        }
      }

      console.log("\n📊 Batch Summary:");
      console.log("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━");
      console.log(`✅ Successfully sent: ${emailsSent}`);
      console.log(`❌ Failed to send: ${emailErrors}`);
      console.log(`📧 Total attempted: ${count}`);
      console.log(
        `📈 Success rate: ${((emailsSent / count) * 100).toFixed(1)}%`
      );
      console.log("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n");

      client.destroy();
      resolve();
    });

    client.on("error", (err) => {
      console.error("\n❌ Connection error:", err);
      resolve();
    });
  });
}

// Interactive menu
async function showMenu() {
  try {
    const template = await loadTemplate();

    console.log("\nGoMailer Test Client");
    console.log("━━━━━━━━━━━━━━━━━━");
    console.log("1. Send single test email");
    console.log("2. Send multiple test emails");
    console.log("3. Exit");
    console.log("━━━━━━━━━━━━━━━━━━");

    rl.question("Select an option: ", async (answer) => {
      switch (answer) {
        case "1":
          await sendSingleEmail(template);
          showMenu();
          break;

        case "2":
          rl.question("Number of emails to send: ", async (count) => {
            await sendTestEmails(template, parseInt(count));
            showMenu();
          });
          break;

        case "3":
          rl.close();
          process.exit(0);
          break;

        default:
          console.log("Invalid option");
          showMenu();
          break;
      }
    });
  } catch (err) {
    console.error("Error loading template:", err);
    process.exit(1);
  }
}

// Start the client
showMenu();
