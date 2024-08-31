import React, { useState, useEffect } from 'react';
import { FaSave, FaPlay, FaTrash } from 'react-icons/fa';

export function SavedCommands({ activeProfile }) {
    const [commands, setCommands] = useState([]);
    const [newCommandName, setNewCommandName] = useState('');
    const [newCommandContent, setNewCommandContent] = useState('');
    const [error, setError] = useState(null);
    const [isExecuting, setIsExecuting] = useState(false);

    useEffect(() => {
        fetchSavedCommands();
    }, []);

    const fetchSavedCommands = async () => {
        try {
            const savedCommands = await window.go.main.App.ListSavedCommands();
            setCommands(savedCommands);
            setError(null);
        } catch (error) {
            console.error('Failed to fetch saved commands:', error);
            setError('Failed to load saved commands. Please try again.');
        }
    };

    const handleSaveCommand = async () => {
        if (!newCommandName || !newCommandContent) return;

        try {
            await window.go.main.App.SaveCommand(newCommandName, newCommandContent);
            fetchSavedCommands();
            setNewCommandName('');
            setNewCommandContent('');
            setError(null);
        } catch (error) {
            console.error('Failed to save command:', error);
            setError('Failed to save command. Please try again.');
        }
    };

    const handleExecuteCommand = async (commandName) => {
        if (!activeProfile) {
            setError('Please connect to a profile before executing commands.');
            return;
        }

        setIsExecuting(true);
        try {
            await window.go.main.App.ExecuteSavedCommand(activeProfile, commandName);
            setError(null);
        } catch (error) {
            console.error('Failed to execute saved command:', error);
            setError('Failed to execute command. Please try again.');
        } finally {
            setIsExecuting(false);
        }
    };

    const handleDeleteCommand = async (name) => {
        try {
            await window.go.main.App.DeleteSavedCommand(name);
            fetchSavedCommands();
            setError(null);
        } catch (error) {
            console.error('Failed to delete saved command:', error);
            setError('Failed to delete command. Please try again.');
        }
    };

    return (
        <div className="mt-4 bg-gray-800 p-4 rounded-lg shadow-md">
            <h2 className="text-xl font-bold mb-4 text-gray-200">Saved Commands</h2>
            {error && <div className="bg-red-500 text-white p-2 mb-4 rounded">{error}</div>}
            <div className="mb-4 flex flex-col space-y-2">
                <input
                    type="text"
                    placeholder="Command Name"
                    value={newCommandName}
                    onChange={(e) => setNewCommandName(e.target.value)}
                    className="p-2 border rounded bg-gray-700 text-white"
                />
                <textarea
                    placeholder="Command Content"
                    value={newCommandContent}
                    onChange={(e) => setNewCommandContent(e.target.value)}
                    className="p-2 border rounded bg-gray-700 text-white h-24"
                />
                <button
                    onClick={handleSaveCommand}
                    className="bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600 transition duration-300 flex items-center justify-center"
                >
                    <FaSave className="mr-2" /> Save Command
                </button>
            </div>
            {commands.length > 0 ? (
                <ul className="space-y-2">
                    {commands.map((command) => (
                        <li key={command.name} className="bg-gray-700 p-3 rounded flex items-center justify-between">
                            <div>
                                <span className="font-bold text-gray-200">{command.name}</span>
                                <p className="text-gray-400 text-sm mt-1">{command.command}</p>
                            </div>
                            <div className="flex space-x-2">
                                <button
                                    onClick={() => handleExecuteCommand(command.name)}
                                    disabled={isExecuting}
                                    className={`bg-blue-500 text-white px-3 py-1 rounded hover:bg-blue-600 transition duration-300 ${isExecuting ? 'opacity-50 cursor-not-allowed' : ''}`}
                                >
                                    <FaPlay />
                                </button>
                                <button
                                    onClick={() => handleDeleteCommand(command.name)}
                                    className="bg-red-500 text-white px-3 py-1 rounded hover:bg-red-600 transition duration-300"
                                >
                                    <FaTrash />
                                </button>
                            </div>
                        </li>
                    ))}
                </ul>
            ) : (
                <p className="text-gray-400">No saved commands found.</p>
            )}
        </div>
    );
}