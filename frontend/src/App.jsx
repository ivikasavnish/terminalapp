import React, { useState, useEffect } from 'react';
import { ProfileSelector } from './profileselector.jsx';
import { Terminal } from './terminal.jsx';
import { PortForwarding } from './portforwarding.jsx';
import { FaTerminal, FaExchangeAlt, FaServer, FaSignOutAlt, FaExclamationTriangle } from 'react-icons/fa';

export default function App() {
    const [activeProfile, setActiveProfile] = useState(null);
    const [connectedProfiles, setConnectedProfiles] = useState([]);
    const [isConnecting, setIsConnecting] = useState(false);
    const [error, setError] = useState(null);

    useEffect(() => {
        fetchConnectedProfiles();
    }, []);

    const fetchConnectedProfiles = async () => {
        try {
            const profiles = await window.go.main.App.GetActiveConnections();
            setConnectedProfiles(profiles || []);
        } catch (error) {
            console.error('Failed to fetch connected profiles:', error);
            setError('Failed to fetch connected profiles. Please try again.');
        }
    };

    const handleConnect = async (profile) => {
        setIsConnecting(true);
        setError(null);
        try {
            const profileJSON = JSON.stringify(profile);
            const result = await window.go.main.App.ConnectSSHWithHostKeyCheck(profileJSON);
            if (result && result.name) {
                setActiveProfile(result.name);
                await fetchConnectedProfiles();
            } else {
                throw new Error("Connection failed: No valid result returned");
            }
        } catch (error) {
            console.error('Failed to connect:', error);
            setError(error.message || "Failed to connect. Please check the server logs for more details.");
        } finally {
            setIsConnecting(false);
        }
    };

    const handleDisconnect = async (profileName) => {
        try {
            await window.go.main.App.DisconnectSSH(profileName);
            setActiveProfile(null);
            await fetchConnectedProfiles();
        } catch (error) {
            console.error('Failed to disconnect:', error);
            setError(error.message || "Failed to disconnect. Please check the server logs for more details.");
        }
    };

    return (
        <div className="flex h-screen bg-gray-900 text-gray-200">
            <aside className="w-1/4 bg-gray-800 p-4 flex flex-col">
                <h1 className="text-2xl font-bold mb-4">SSH Client</h1>
                <ProfileSelector
                    activeProfile={activeProfile}
                    connectedProfiles={connectedProfiles}
                    onConnect={handleConnect}
                    onDisconnect={handleDisconnect}
                    isConnecting={isConnecting}
                />
            </aside>
            <main className="flex-1 flex flex-col">
                {error && (
                    <div className="bg-red-500 text-white p-2 flex items-center">
                        <FaExclamationTriangle className="mr-2" />
                        {error}
                    </div>
                )}
                {activeProfile ? (
                    <>
                        <header className="bg-gray-700 p-4 flex justify-between items-center">
                            <div className="flex items-center">
                                <FaServer className="mr-2" />
                                <span className="font-semibold">{activeProfile}</span>
                            </div>
                            <button
                                onClick={() => handleDisconnect(activeProfile)}
                                className="bg-red-500 text-white px-4 py-2 rounded hover:bg-red-600 transition duration-200 flex items-center"
                            >
                                <FaSignOutAlt className="mr-2" />
                                Disconnect
                            </button>
                        </header>
                        <div className="flex-1 flex">
                            <div className="w-2/3 border-r border-gray-700">
                                <div className="bg-gray-800 p-2">
                                    <h2 className="text-lg font-semibold flex items-center">
                                        <FaTerminal className="mr-2" />
                                        Terminal
                                    </h2>
                                </div>
                                <Terminal activeProfile={activeProfile} />
                            </div>
                            <div className="w-1/3">
                                <div className="bg-gray-800 p-2">
                                    <h2 className="text-lg font-semibold flex items-center">
                                        <FaExchangeAlt className="mr-2" />
                                        Port Forwarding
                                    </h2>
                                </div>
                                <PortForwarding activeProfile={activeProfile} />
                            </div>
                        </div>
                    </>
                ) : (
                    <div className="flex-1 flex items-center justify-center">
                        <p className="text-xl text-gray-400">Please select a profile to connect.</p>
                    </div>
                )}
            </main>
        </div>
    );
}