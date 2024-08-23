import React, { useState, useEffect } from 'react';
import { FaPlus, FaMinus, FaDesktop, FaServer, FaUserPlus, FaSpinner } from 'react-icons/fa';

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

export function ProfileSelector({ activeProfile, connectedProfiles = [], onConnect, onDisconnect, isConnecting }) {
    const [profiles, setProfiles] = useState({ base: null, standard: [], custom: [] });
    const [error, setError] = useState(null);

    useEffect(() => {
        fetchProfiles();
    }, []);

    const fetchProfiles = async () => {
        try {
            const [baseProfile, standardProfiles] = await Promise.all([
                window.go.main.App.GetBaseProfile(),
                window.go.main.App.LoadProfiles(),
            ]);
            setProfiles({
                base: baseProfile,
                standard: standardProfiles,
                custom: [], // Implement custom profiles if needed
            });
        } catch (error) {
            console.error('Failed to fetch profiles:', error);
            setError('Failed to fetch profiles. Please try again.');
        }
    };

    const renderProfileItem = (profile, type) => {
        if (!profile) return null;
        const profileName = profile.name || profile.host || `Unnamed ${type} Profile`;
        const isConnected = Array.isArray(connectedProfiles) && connectedProfiles.includes(profileName);
        const Icon = ProfileIcons[type];

        return (
            <li key={profileName} className="flex items-center justify-between mb-2 bg-gray-700 p-2 rounded">
                <span className="flex items-center">
                    <Icon className="mr-2 text-blue-400" />
                    <div>
                        <div>{profileName}</div>
                        <div className="text-sm text-gray-400">{profile.username}@{profile.host}:{profile.port}</div>
                    </div>
                </span>
                <button
                    onClick={() => isConnected ? onDisconnect(profileName) : onConnect(profile)}
                    className={`${isConnected ? 'bg-red-500' : 'bg-green-500'} text-white px-2 py-1 rounded flex items-center`}
                    disabled={isConnecting}
                >
                    {isConnecting ? (
                        <FaSpinner className="mr-1 animate-spin" />
                    ) : isConnected ? (
                        <><FaMinus className="mr-1" /> Disconnect</>
                    ) : (
                        <><FaPlus className="mr-1" /> Connect</>
                    )}
                </button>
            </li>
        );
    };

    return (
        <div className="mb-4">
            {error && <div className="bg-red-500 text-white p-2 rounded mb-4">{error}</div>}
            <h2 className="text-xl font-bold mb-2">Base Profile</h2>
            {profiles.base && renderProfileItem(profiles.base, ProfileTypes.BASE)}
            <h2 className="text-xl font-bold mb-2 mt-4">Standard Profiles</h2>
            {profiles.standard.length > 0 ? (
                <ul className="mb-4">
                    {profiles.standard.map(profile => renderProfileItem(profile, ProfileTypes.STANDARD))}
                </ul>
            ) : (
                <p className="text-gray-400">No standard profiles available.</p>
            )}
        </div>
    );
}