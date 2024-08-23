import React, { useState, useEffect } from 'react';
import { FaSave, FaPlay, FaTrash } from 'react-icons/fa';

export function SavedCommands({ activeProfile }) {
    const [commands, setCommands] = useState([]);
    const [newCommandName, setNewCommandName] = useState('');
    const [newCommandContent, setNewCommandContent] = useState('');
    const [error, setError] = useState(null);

    useEffect(() => {
        fetchSavedCommands();
    }, [activeProfile]);

    const fetchSavedCommands = async () => {
        try {
            const savedCommands = await window.go.main.App.ListSavedCommands();
            if (Array.isArray(savedCommands)) {
                setCommands(savedCommands);
            } else if (typeof savedCommands === 'object') {
                // If it's an object, convert it to an array
                setCommands(Object.entries(savedCommands).map(([name, command]) => ({ Name: name, Command: command })));
            } else {
                throw new Error('Unexpected data structure for saved commands');
            }
            setError(null);
        } catch (error) {
            console.error('Failed to fetch saved commands:', error);
            setError('Failed to load saved commands. Please try again.');
            setCommands([]);
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
        try {
            const result = await window.go.main.App.ExecuteSavedCommand(activeProfile, commandName);
            console.log('Command execution result:', result);
            // You might want to update the Terminal component with this result
            // or display it in some way to the user
            setError(null);
        } catch (error) {
            console.error('Failed to execute saved command:', error);
            setError('Failed to execute command. Please try again.');
        }
    };

    const handleDeleteCommand = async (commandName) => {
        try {
            await window.go.main.App.DeleteSavedCommand(commandName);
            fetchSavedCommands();
            setError(null);
        } catch (error) {
            console.error('Failed to delete saved command:', error);
            setError('Failed to delete command. Please try again.');
        }
    };

    return (
        <div className="mt-4">
            <h2 className="text-xl font-bold mb-2">Saved Commands</h2>
            {error && <div className="text-red-500 mb-2">{error}</div>}
            <div className="mb-4">
                <input
                    type="text"
                    placeholder="Command Name"
                    value={newCommandName}
                    onChange={(e) => setNewCommandName(e.target.value)}
                    className="mr-2 p-1 border"
                />
                <input
                    type="text"
                    placeholder="Command Content"
                    value={newCommandContent}
                    onChange={(e) => setNewCommandContent(e.target.value)}
                    className="mr-2 p-1 border"
                />
                <button
                    onClick={handleSaveCommand}
                    className="bg-green-500 text-white px-2 py-1 rounded"
                >
                    <FaSave className="inline mr-1" /> Save Command
                </button>
            </div>
            {Array.isArray(commands) && commands.length > 0 ? (
                <ul>
                    {commands.map((command) => (
                        <li key={command.Name} className="flex items-center justify-between mb-2 p-2 bg-gray-100 rounded">
              <span className="mr-2">
                <strong>{command.Name}:</strong> {command.Command}
              </span>
                            <div>
                                <button
                                    onClick={() => handleExecuteCommand(command.Name)}
                                    className="bg-blue-500 text-white px-2 py-1 rounded mr-2"
                                >
                                    <FaPlay className="inline mr-1" /> Execute
                                </button>
                                <button
                                    onClick={() => handleDeleteCommand(command.Name)}
                                    className="bg-red-500 text-white px-2 py-1 rounded"
                                >
                                    <FaTrash className="inline mr-1" /> Delete
                                </button>
                            </div>
                        </li>
                    ))}
                </ul>
            ) : (
                <p>No saved commands found.</p>
            )}
        </div>
    );
}