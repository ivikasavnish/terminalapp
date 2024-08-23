import React, { useState, useEffect } from 'react';
import { FaPlus, FaMinus, FaTrash, FaCheck, FaTimes, FaDesktop, FaServer, FaUserPlus } from 'react-icons/fa';

const ProfileTypes = {
    BASE: 'base',
    STANDARD: 'standard',
    CUSTOM: 'custom',
};

const ProfileIcons = {
    [ProfileTypes.BASE]: FaDesktop,
    [ProfileTypes.STANDARD]: FaServer,
    [ProfileTypes.CUSTOM]: FaUserPlus,
};

export function ProfileSelector({ activeProfile, connectedProfiles, onConnect, onDisconnect }) {
    const [profiles, setProfiles] = useState({ base: null, standard: [], custom: [] });
    const [showManualConnect, setShowManualConnect] = useState(false);
    const [manualConnection, setManualConnection] = useState({
        name: '', username: '', password: '', host: '', port: '22'
    });
    const [error, setError] = useState(null);
    const [isConnecting, setIsConnecting] = useState(false);

    useEffect(() => {
        fetchProfiles();
    }, []);

    const fetchProfiles = async () => {
        try {
            const [baseProfile, standardProfiles, customProfiles] = await Promise.all([
                window.go.main.App.GetBaseProfile(),
                window.go.main.App.LoadProfiles(),
                window.go.main.App.LoadCustomProfiles()
            ]);
            setProfiles({
                base: baseProfile || null,
                standard: Array.isArray(standardProfiles) ? standardProfiles : [],
                custom: Array.isArray(customProfiles) ? customProfiles : []
            });
        } catch (error) {
            console.error('Failed to fetch profiles:', error);
            setError('Failed to fetch profiles. Please try again.');
        }
        console.log(profiles);
    };

    const handleConnect = async (profile, type) => {
        setIsConnecting(true);
        setError(null);
        try {
            console.log("Connecting with profile:", profile);
            const result = await window.go.main.App.ConnectSSHWithHostKeyCheck(profile);
            if (result) {
                throw new Error(result);
            }
            onConnect(profile.name || profile.host);
            console.log("Connected successfully");
        } catch (error) {
            console.error("Connection error:", error);
            setError(`Failed to connect: ${error.message || 'Unknown error'}`);
        } finally {
            setIsConnecting(false);
        }
    };

    const handleDisconnect = async (profile) => {
        try {
            await window.go.main.App.DisconnectSSH(profile.name || profile.host);
            onDisconnect(profile.name || profile.host);
            console.log("Disconnected successfully");
        } catch (error) {
            console.error("Disconnection error:", error);
            setError(`Failed to disconnect: ${error.message || 'Unknown error'}`);
        }
    };

    const handleManualConnect = async () => {
        await handleConnect(manualConnection);
        setShowManualConnect(false);
        setManualConnection({ name: '', username: '', password: '', host: '', port: '22' });
    };

    const renderProfileItem = (profile, type) => {
        if (!profile) return null;
        let profileName, profileDetails;

        if (type === ProfileTypes.BASE) {
            profileName = profile.name || "Host System";
            profileDetails = `${profile.username || ''}@${profile.host || ''}:${profile.port || ''}`;
        } else if (type === ProfileTypes.STANDARD || type === ProfileTypes.CUSTOM) {
            profileName = profile.name || '';
            profileDetails = `${profile.username || ''}@${profile.host || ''}:${profile.port || ''}`;
        } else {
            return null; // Invalid profile type
        }
    const renderProfileSection = (title, profileType) => {
        let profilesToRender = profileType === ProfileTypes.BASE ? (profiles.base ? [profiles.base] : []) : (profiles[profileType] || []);

        return (
            <>
                <h2 className="text-xl font-bold mb-2">{title}</h2>
                {profilesToRender.length > 0 ? (
                    <ul className="mb-4">
                        {profilesToRender.map((profile, index) => renderProfileItem(profile, profileType))}
                    </ul>
                ) : (
                    <p className="mb-4 text-gray-400">No {profileType} profiles available.</p>
                )}
            </>
        );
    };

    return (
        <div className="mb-4">
            {error && (
                <div className="bg-red-500 text-white p-2 rounded mb-4 flex items-center justify-between">
                    <span>{error}</span>
                    <button onClick={() => setError(null)} className="text-white"><FaTimes /></button>
                </div>
            )}
            {renderProfileSection('Base Profile', ProfileTypes.BASE)}
            {renderProfileSection('Profiles', ProfileTypes.STANDARD)}
            {renderProfileSection('Custom Profiles', ProfileTypes.CUSTOM)}
            <button
                onClick={() => setShowManualConnect(!showManualConnect)}
                className="w-full bg-blue-500 text-white px-2 py-1 rounded mb-4"
            >
                {showManualConnect ? 'Hide Manual Connect' : 'Manual Connect'}
            </button>
            {showManualConnect && (
                <div className="bg-gray-700 p-4 rounded">
                    {['name', 'username', 'password', 'host', 'port'].map(field => (
                        <input
                            key={field}
                            type={field === 'password' ? 'password' : 'text'}
                            placeholder={field.charAt(0).toUpperCase() + field.slice(1)}
                            value={manualConnection[field]}
                            onChange={(e) => setManualConnection({...manualConnection, [field]: e.target.value})}
                            className="mb-2 w-full p-2 bg-gray-600 rounded"
                        />
                    ))}
                    <div className="flex justify-between">
                        <button
                            onClick={handleManualConnect}
                            className="bg-green-500 text-white px-2 py-1 rounded flex items-center"
                            disabled={isConnecting}
                        >
                            <FaPlus className="mr-1" /> Connect
                        </button>
                    </div>
                </div>
            )}
        </div>
    );
}