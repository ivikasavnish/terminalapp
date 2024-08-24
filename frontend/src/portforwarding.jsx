import React, { useState, useEffect } from 'react';
import { FaPlay, FaStop } from 'react-icons/fa';

export function PortForwarding({ activeProfile }) {
    const [localPort, setLocalPort] = useState('');
    const [remotePort, setRemotePort] = useState('');
    const [isRemoteToLocal, setIsRemoteToLocal] = useState(false);
    const [status, setStatus] = useState('');
    const [activeForwards, setActiveForwards] = useState([]);

    useEffect(() => {
        if (activeProfile) {
            fetchActiveForwards();
        } else {
            setActiveForwards([]);
        }
    }, [activeProfile]);

    const fetchActiveForwards = async () => {
        try {
            const forwards = await window.go.main.App.GetActivePortForwards(activeProfile);
            setActiveForwards(forwards);
        } catch (error) {
            console.error('Failed to fetch active port forwards:', error);
            setStatus('Failed to fetch active port forwards. Please try again.');
        }
    };

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
            setStatus('Port forwarding set up successfully.');
            fetchActiveForwards();
            setLocalPort('');
            setRemotePort('');
        } catch (error) {
            console.error('Failed to set up port forwarding:', error);
            setStatus(`Failed to set up port forwarding: ${error.message}`);
        }
    };

    const handleStopForward = async (localPort, remotePort, isRemoteToLocal) => {
        try {
            await window.go.main.App.StopPortForward(
                activeProfile,
                localPort,
                remotePort,
                isRemoteToLocal
            );
            setStatus('Port forwarding stopped successfully.');
            fetchActiveForwards();
        } catch (error) {
            console.error('Failed to stop port forwarding:', error);
            setStatus(`Failed to stop port forwarding: ${error.message}`);
        }
    };

    return (
        <div className="bg-gray-800 p-4 rounded-lg shadow-md">
            <h2 className="text-xl font-bold mb-4 text-gray-200">Port Forwarding</h2>
            <div className="mb-4 grid grid-cols-2 gap-4">
                <input
                    type="number"
                    placeholder="Local Port"
                    value={localPort}
                    onChange={(e) => setLocalPort(e.target.value)}
                    className="w-full p-2 bg-gray-700 text-white rounded"
                />
                <input
                    type="number"
                    placeholder="Remote Port"
                    value={remotePort}
                    onChange={(e) => setRemotePort(e.target.value)}
                    className="w-full p-2 bg-gray-700 text-white rounded"
                />
            </div>
            <div className="mb-4 flex items-center">
                <label className="flex items-center text-gray-300">
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
                    className="ml-auto bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600 transition duration-200"
                >
                    <FaPlay className="inline-block mr-2" />
                    Set Up Port Forward
                </button>
            </div>
            {status && (
                <div className={`p-2 rounded ${status.includes('successfully') ? 'bg-green-600' : 'bg-red-600'} text-white mb-4`}>
                    {status}
                </div>
            )}
            <div className="mb-4">
                <h3 className="font-bold text-gray-200 mb-2">Active Port Forwards:</h3>
                {activeForwards.length === 0 ? (
                    <p className="text-gray-400">No active port forwards.</p>
                ) : (
                    activeForwards.map((forward, index) => (
                        <div key={index} className="flex justify-between items-center bg-gray-700 p-2 rounded mt-2">
                            <span className="text-gray-300">
                                {forward.isRemoteToLocal ? 'Remote' : 'Local'} {forward.localPort} →
                                {forward.isRemoteToLocal ? 'Local' : 'Remote'} {forward.remotePort}
                            </span>
                            <button
                                onClick={() => handleStopForward(forward.localPort, forward.remotePort, forward.isRemoteToLocal)}
                                className="bg-red-500 text-white px-2 py-1 rounded hover:bg-red-600 transition duration-200"
                            >
                                <FaStop className="inline-block mr-1" /> Stop
                            </button>
                        </div>
                    ))
                )}
            </div>
        </div>
    );
}