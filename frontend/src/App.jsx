import React, { useState, useEffect } from 'react';
import { ProfileSelector } from './profileselector.jsx';

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
            console.log('Attempting to connect with profile:', profile);
            const profileJSON = JSON.stringify(profile);
            const result = await window.go.main.App.ConnectSSHWithHostKeyCheck(profileJSON);
            console.log('Connection result:', result);
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
        <div className="flex flex-col h-screen bg-gray-900 text-gray-200">
            <header className="bg-blue-600 text-white p-4">
                <h1 className="text-2xl font-bold">SSH Client</h1>
            </header>
            <div className="flex-grow flex overflow-hidden">
                <aside className="w-1/4 bg-gray-800 p-4 overflow-y-auto">
                    {error && (
                        <div className="bg-red-500 text-white p-2 rounded mb-4">
                            {error}
                        </div>
                    )}
                    <ProfileSelector
                        activeProfile={activeProfile}
                        connectedProfiles={connectedProfiles}
                        onConnect={handleConnect}
                        onDisconnect={handleDisconnect}
                        isConnecting={isConnecting}
                    />
                </aside>
                <main className="w-3/4 flex flex-col overflow-hidden">
                    {activeProfile ? (
                        <div className="flex-grow overflow-hidden">
                            <p>Connected to: {activeProfile}</p>
                            {/* Add your terminal component here */}
                        </div>
                    ) : (
                        <div className="flex items-center justify-center h-full">
                            <p className="text-xl text-gray-400">Please select a profile to connect.</p>
                        </div>
                    )}
                </main>
            </div>
        </div>
    );
}