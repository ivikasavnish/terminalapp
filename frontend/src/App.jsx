import React, { useState, useEffect } from 'react';
import { ProfileSelector } from './ProfileSelector';
import { Terminal } from './Terminal';
import { PortForwarding } from './PortForwarding';
import { FileBrowser } from './FileBrowser';
import { SavedCommands } from './SavedCommands';
import { FaTerminal, FaExchangeAlt, FaServer, FaSignOutAlt, FaExclamationTriangle, FaFolder } from 'react-icons/fa';

export default function App() {
    const [activeProfile, setActiveProfile] = useState(null);
    const [connectedProfiles, setConnectedProfiles] = useState([]);
    const [isConnecting, setIsConnecting] = useState(false);
    const [error, setError] = useState(null);
    const [activeTab, setActiveTab] = useState('terminal');

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
                <SavedCommands activeProfile={activeProfile} />
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
                            <div className="flex">
                                <button
                                    onClick={() => setActiveTab('terminal')}
                                    className={`mr-2 px-4 py-2 rounded ${activeTab === 'terminal' ? 'bg-blue-500' : 'bg-gray-600'}`}
                                >
                                    <FaTerminal className="mr-2 inline" />
                                    Terminal
                                </button>
                                <button
                                    onClick={() => setActiveTab('portForwarding')}
                                    className={`mr-2 px-4 py-2 rounded ${activeTab === 'portForwarding' ? 'bg-blue-500' : 'bg-gray-600'}`}
                                >
                                    <FaExchangeAlt className="mr-2 inline" />
                                    Port Forwarding
                                </button>
                                <button
                                    onClick={() => setActiveTab('fileBrowser')}
                                    className={`mr-2 px-4 py-2 rounded ${activeTab === 'fileBrowser' ? 'bg-blue-500' : 'bg-gray-600'}`}
                                >
                                    <FaFolder className="mr-2 inline" />
                                    File Browser
                                </button>
                                <button
                                    onClick={() => handleDisconnect(activeProfile)}
                                    className="bg-red-500 text-white px-4 py-2 rounded hover:bg-red-600 transition duration-200 flex items-center"
                                >
                                    <FaSignOutAlt className="mr-2" />
                                    Disconnect
                                </button>
                            </div>
                        </header>
                        <div className="flex-1 overflow-hidden">
                            {activeTab === 'terminal' && <Terminal activeProfile={activeProfile} />}
                            {activeTab === 'portForwarding' && <PortForwarding activeProfile={activeProfile} />}
                            {activeTab === 'fileBrowser' && <FileBrowser activeProfile={activeProfile} />}
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