import React, { useState, useEffect } from 'react';
import { ProfileSelector } from './profileselector.jsx';
// import { Terminal } from './terminal.jsx';
// import { FileBrowser } from './filebrowser.jsx';
// import { PortForwarding } from './portforwarding.jsx';
// import { SavedCommands } from './savedCommands.jsx';

export default function App() {
    const [activeProfile, setActiveProfile] = useState(null);
    const [connectedProfiles, setConnectedProfiles] = useState([]);

    useEffect(() => {
        fetchConnectedProfiles();
    }, []);

    const fetchConnectedProfiles = async () => {
        try {
            const profiles = await window.go.main.App.GetActiveConnections();
            setConnectedProfiles(profiles);
        } catch (error) {
            console.error('Failed to fetch connected profiles:', error);
        }
    };

    const handleConnect = async (profile) => {
        try {
            await window.go.main.App.ConnectSSH(profile);
            setActiveProfile(profile);
            fetchConnectedProfiles();
        } catch (error) {
            console.error('Failed to connect:', error);
        }
    };

    const handleDisconnect = async (profile) => {
        try {
            await window.go.main.App.DisconnectSSH(profile);
            setActiveProfile(null);
            fetchConnectedProfiles();
        } catch (error) {
            console.error('Failed to disconnect:', error);
        }
    };

    return (
        <div className="flex flex-col h-screen bg-gray-900 text-gray-200">
            <header className="bg-blue-600 text-white p-4">
                <h1 className="text-2xl font-bold">SSH Client</h1>
            </header>
            <div className="flex-grow flex overflow-hidden">
                <aside className="w-1/4 bg-gray-800 p-4 overflow-y-auto">
                    <ProfileSelector
                        activeProfile={activeProfile}
                        connectedProfiles={connectedProfiles}
                        onConnect={handleConnect}
                        onDisconnect={handleDisconnect}
                    />
                    {/*<SavedCommands activeProfile={activeProfile} />*/}
                </aside>
                <main className="w-3/4 flex flex-col overflow-hidden">
                    {activeProfile ? (
                        <>
                            <div className="flex-grow overflow-hidden">
                                {/*<Terminal activeProfile={activeProfile} />*/}
                            </div>
                            <div className="h-1/3 flex border-t border-gray-700">
                                {/*<FileBrowser activeProfile={activeProfile} />*/}
                                {/*<PortForwarding activeProfile={activeProfile} />*/}
                            </div>
                        </>
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