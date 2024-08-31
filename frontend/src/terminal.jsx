import React, { useState, useEffect, useRef } from 'react';
import { FaPlay, FaStop } from 'react-icons/fa';
import cloudCommands from './cloudcommands.js';

export function Terminal({ activeProfile }) {
    const [output, setOutput] = useState([]);
    const [input, setInput] = useState('');
    const [isRunning, setIsRunning] = useState(false);
    const [suggestions, setSuggestions] = useState([]);
    const [selectedSuggestion, setSelectedSuggestion] = useState(-1);
    const outputRef = useRef(null);
    const inputRef = useRef(null);

    useEffect(() => {
        const handleCommandOutput = (event) => {
            if (event.profile === activeProfile) {
                setOutput(prev => [...prev, { type: event.type, content: event.data }]);
                if (event.type === 'info' && (event.data.includes('Command finished') || event.data.includes('Command stopped'))) {
                    setIsRunning(false);
                }
                if (event.type === 'error') {
                    setIsRunning(false);
                }
            }
        };

        window.runtime.EventsOn("command_output", handleCommandOutput);

        return () => {
            window.runtime.EventsOff("command_output", handleCommandOutput);
        };
    }, [activeProfile]);

    useEffect(() => {
        if (outputRef.current) {
            outputRef.current.scrollTop = outputRef.current.scrollHeight;
        }
    }, [output]);

    const handleInputChange = (e) => {
        const value = e.target.value;
        setInput(value);
        updateSuggestions(value);
    };

    const updateSuggestions = (value) => {
        const inputParts = value.split(' ');
        const currentWord = inputParts[inputParts.length - 1].toLowerCase();

        let newSuggestions = [];

        // Suggest commands
        if (inputParts.length === 1) {
            newSuggestions = [...cloudCommands.gcp, ...cloudCommands.aws]
                .filter(cmd => cmd.command.toLowerCase().startsWith(currentWord))
                .map(cmd => cmd.command);
        }
        // Suggest arguments
        else {
            const fullCommand = inputParts.slice(0, -1).join(' ');
            const command = [...cloudCommands.gcp, ...cloudCommands.aws].find(cmd => cmd.command === fullCommand);
            if (command) {
                newSuggestions = command.args.filter(arg => arg.toLowerCase().startsWith(currentWord));
            }
        }

        setSuggestions(newSuggestions);
        setSelectedSuggestion(-1);
    };

    const handleKeyDown = (e) => {
        if (e.key === 'Tab' && suggestions.length > 0) {
            e.preventDefault();
            const newInput = input.split(' ');
            newInput[newInput.length - 1] = suggestions[0];
            setInput(newInput.join(' '));
            setSuggestions([]);
        } else if (e.key === 'ArrowUp' || e.key === 'ArrowDown') {
            e.preventDefault();
            if (suggestions.length > 0) {
                const direction = e.key === 'ArrowUp' ? -1 : 1;
                const newIndex = (selectedSuggestion + direction + suggestions.length) % suggestions.length;
                setSelectedSuggestion(newIndex);
            }
        } else if (e.key === 'Enter') {
            if (selectedSuggestion !== -1) {
                const newInput = input.split(' ');
                newInput[newInput.length - 1] = suggestions[selectedSuggestion];
                setInput(newInput.join(' '));
                setSuggestions([]);
                setSelectedSuggestion(-1);
            } else {
                executeCommand();
            }
        }
    };

    const executeCommand = async () => {
        if (!input.trim() || isRunning) return;

        setOutput(prev => [...prev, { type: 'input', content: input }]);
        setIsRunning(true);

        try {
            await window.go.main.App.ExecuteInteractiveCommand(activeProfile, input);
        } catch (error) {
            console.error('Failed to execute command:', error);
            setOutput(prev => [...prev, { type: 'error', content: `Failed to execute command: ${error.message}` }]);
            setIsRunning(false);
        }

        setInput('');
        setSuggestions([]);
    };

    const stopCommand = async () => {
        try {
            await window.go.main.App.StopInteractiveCommand(activeProfile);
        } catch (error) {
            console.error('Failed to stop command:', error);
            setOutput(prev => [...prev, { type: 'error', content: `Failed to stop command: ${error.message}` }]);
        }
    };

    return (
        <div className="flex flex-col h-full bg-gray-900 text-gray-300 font-mono">
            <div ref={outputRef} className="flex-grow overflow-auto p-2">
                {output.map((item, index) => (
                    <div key={index} className={`mb-1 ${
                        item.type === 'input' ? 'text-blue-400' :
                            item.type === 'error' ? 'text-red-400' :
                                item.type === 'stderr' ? 'text-yellow-400' :
                                    'text-gray-300'
                    }`}>
                        {item.type === 'input' ? '$ ' : ''}{item.content}
                    </div>
                ))}
            </div>
            <div className="relative p-2 border-t border-gray-700">
                <input
                    ref={inputRef}
                    type="text"
                    value={input}
                    onChange={handleInputChange}
                    onKeyDown={handleKeyDown}
                    className="w-full bg-gray-800 text-gray-300 px-2 py-1 rounded focus:outline-none"
                    placeholder={isRunning ? "Command is running..." : "Enter command..."}
                    disabled={isRunning}
                />
                {suggestions.length > 0 && (
                    <ul className="absolute left-0 right-0 bottom-full bg-gray-700 rounded-t overflow-hidden">
                        {suggestions.map((suggestion, index) => (
                            <li
                                key={index}
                                className={`px-2 py-1 ${index === selectedSuggestion ? 'bg-blue-600' : 'hover:bg-gray-600'}`}
                                onClick={() => {
                                    const newInput = input.split(' ');
                                    newInput[newInput.length - 1] = suggestion;
                                    setInput(newInput.join(' '));
                                    setSuggestions([]);
                                    inputRef.current.focus();
                                }}
                            >
                                {suggestion}
                            </li>
                        ))}
                    </ul>
                )}
                {isRunning ? (
                    <button
                        onClick={stopCommand}
                        className="absolute right-2 top-2 bg-red-500 text-white px-2 py-1 rounded hover:bg-red-600 transition duration-200"
                    >
                        <FaStop className="mr-1 inline" /> Stop
                    </button>
                ) : (
                    <button
                        onClick={executeCommand}
                        className="absolute right-2 top-2 bg-green-500 text-white px-2 py-1 rounded hover:bg-green-600 transition duration-200"
                    >
                        <FaPlay className="mr-1 inline" /> Run
                    </button>
                )}
            </div>
        </div>
    );
}