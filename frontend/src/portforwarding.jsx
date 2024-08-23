import React, { useState } from 'react';

export function PortForwarding({ activeProfile }) {
    const [localPort, setLocalPort] = useState('');
    const [remotePort, setRemotePort] = useState('');
    const [isRemoteToLocal, setIsRemoteToLocal] = useState(false);
    const [status, setStatus] = useState('');

    const handlePortForward = async () => {
        if (!localPort || !remotePort) {
            setStatus('Please enter both local and remote ports.');
            return;
        }

        setStatus('Setting up port forwarding...');

        try {
            await window.go.main.App.PortForward(
                activeProfile,
                parseInt(localPort),
                parseInt(remotePort),
                isRemoteToLocal
            );
            if (!isRemoteToLocal) {
                setStatus(`Port forwarding set up successfully. Your local service on port ${localPort} is now accessible via the remote server's IP address on port ${remotePort}.`);
            } else {
                setStatus(`Port forwarding set up successfully. The remote server's port ${remotePort} is now accessible locally on port ${localPort}.`);
            }
        } catch (error) {
            setStatus(`Failed to set up port forwarding: ${error.message}`);
        }
    };

    return (
        <div className="w-full p-4 bg-gray-800 rounded-lg shadow-md">
            <h2 className="text-xl font-bold mb-4 text-gray-200">Port Forwarding</h2>
            <div className="mb-4">
                <input
                    type="number"
                    placeholder="Local Port"
                    value={localPort}
                    onChange={(e) => setLocalPort(e.target.value)}
                    className="w-full mb-2 p-2 bg-gray-700 text-white rounded"
                />
                <input
                    type="number"
                    placeholder="Remote Port"
                    value={remotePort}
                    onChange={(e) => setRemotePort(e.target.value)}
                    className="w-full mb-2 p-2 bg-gray-700 text-white rounded"
                />
                <label className="flex items-center text-gray-300 mb-2">
                    <input
                        type="checkbox"
                        checked={isRemoteToLocal}
                        onChange={(e) => setIsRemoteToLocal(e.target.checked)}
                        className="mr-2"
                    />
                    Remote to Local
                </label>
                <button
                    onClick={handlePortForward}
                    className="w-full bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600 transition duration-200"
                >
                    Set Up Port Forwarding
                </button>
            </div>
            {status && (
                <div className={`p-2 rounded ${status.includes('successfully') ? 'bg-green-600' : 'bg-red-600'} text-white`}>
                    {status}
                </div>
            )}
            <div className="mt-4 text-gray-300">
                <h3 className="font-bold">How to use:</h3>
                <p>1. To expose a local service to the internet:</p>
                <ul className="list-disc list-inside ml-4">
                    <li>Enter your local port (e.g., 8090) in "Local Port"</li>
                    <li>Enter the desired remote port (e.g., 8090) in "Remote Port"</li>
                    <li>Ensure "Remote to Local" is unchecked</li>
                    <li>Click "Set Up Port Forwarding"</li>
                    <li>Your local service will be accessible via the remote server's IP address on the specified remote port</li>
                </ul>
                <p className="mt-2">2. To access a remote service locally:</p>
                <ul className="list-disc list-inside ml-4">
                    <li>Enter the desired local port in "Local Port"</li>
                    <li>Enter the remote service's port in "Remote Port"</li>
                    <li>Check "Remote to Local"</li>
                    <li>Click "Set Up Port Forwarding"</li>
                    <li>The remote service will be accessible on your local machine at the specified local port</li>
                </ul>
            </div>
        </div>
    );
}