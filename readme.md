# Time-Based One-Time Port Knocking (TOPK)

TOPK is a secure and dynamic port knocking solution designed to enhance the security of servers by utilizing time-based one-time sequences for authentication. This approach ensures that port knocking sequences are not only unique to each user but also vary with time, making it significantly harder for potential attackers to gain unauthorized access through sequence prediction or replay attacks.

## Features

- **Dynamic Sequences**: Generates unique port knocking sequences based on the current time and a random hash, ensuring that each sequence is only valid for a short period.
- **User-Specific Sequences**: Supports generating different sequences for different users, allowing for personalized access control.
- **Enhanced Security**: By using time-based sequences and requiring sequences to be completed within a specified timeout, TOPK prevents replay and brute-force attacks more effectively than traditional static port knocking.
- **Flexibility**: Designed to be adaptable for various network configurations and security requirements.
- **Crypto-Random Security**: Utilizes the crypto/rand package for secure random number generation, ensuring the robustness of the security mechanism.

## How It Works

TOPK operates by listening for connection attempts (UDP/TCP) on a series of dynamically generated ports. These ports are determined based on the current time and a cryptographic hash, creating a sequence that must be hit in the correct order for access to be granted. The sequence is unique per user and changes periodically, ensuring that only authenticated users who know the current sequence can access the system.

## Getting Started

### Requirements

- A Linux-based server with firewall rules configured to drop all incoming connections except those necessary for TOPK.
- Go programming environment for building the tool.

##Installation

1.  **Clone the repository:**

```bash
git clone https://github.com/Canadadry/topk.git
cd TOPK
```

2.  **Build the tool:**

```bash
go build -o topk
```

3.  **Configure the tool (optional):**

Before running TOPK, you may want to configure user-specific sequences or adjust the timeout settings based on your security requirements. Edit the configuration file config.json accordingly.

4.  **Run TOPK:**

```bash
sudo ./topk
```

Ensure you run TOPK with root privileges as it needs to listen on network interfaces and modify firewall rules dynamically.

## Usage

After starting TOPK, it will begin listening for port knocking sequences. Users must "knock" on the dynamically determined ports in the correct order within the allowed time frame to gain access.

1.  **Determine your current sequence:**
    Use the provided client tool or script to calculate your current sequence based on the shared secret and the current time.

2.  **Send the sequence:**
    Knock on the ports in the determined sequence from your client machine.
3.  **Access the server:**
    If the sequence is correct and completed within the allowed time frame, TOPK will temporarily open the firewall, allowing you to initiate a connection to the server.

## Security Considerations

- Ensure that the time is synchronized between the server and client machines using NTP to avoid sequence mismatches due to time discrepancies.
- Keep the hash secret shared between the server and authorized users secure to prevent unauthorized access.
- Regularly monitor server logs for any unauthorized access attempts or unusual activity.

## Contributing

Contributions to TOPK are welcome! Please refer to the contributing guidelines for more information on how to submit pull requests, report issues, or request features.

## License

TOPK is licensed under the MIT License. See the LICENSE file for more details.
