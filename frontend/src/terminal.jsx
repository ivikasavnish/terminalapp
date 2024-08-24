import React, { useState, useEffect, useRef } from 'react';
import { FaPlay, FaStop, FaTrash } from 'react-icons/fa';

export function Terminal({ activeProfile }) {
    const [output, setOutput] = useState([]);
    const [input, setInput] = useState('');
    const [isRunning, setIsRunning] = useState(false);
    const [commandHistory, setCommandHistory] = useState([]);
    const [historyIndex, setHistoryIndex] = useState(-1);
    const outputRef = useRef(null);
    const inputRef = useRef(null);

    useEffect(() => {
        const handleCommandOutput = (event) => {
            if (event.profile === activeProfile) {
                if (event.type === 'clear') {
                    setOutput([]);
                } else {
                    setOutput(prev => [...prev, { type: event.type, content: event.data }]);
                }
                if (event.type === 'info' && (event.data.includes('Command finished') || event.data.includes('Command stopped'))) {
                    setIsRunning(false);
                }
                if (event.type === 'error') {
                    setIsRunning(false);
                }
            }
        };

        window.runtime.EventsOn("command_output", handleCommandOutput);

        loadCommandHistory();

        return () => {
            window.runtime.EventsOff("command_output", handleCommandOutput);
        };
    }, [activeProfile]);

    useEffect(() => {
        if (outputRef.current) {
            outputRef.current.scrollTop = outputRef.current.scrollHeight;
        }
    }, [output]);

    const loadCommandHistory = async () => {
        try {
            const history = await window.go.main.App.GetCommandHistory(activeProfile);
            setCommandHistory(history);
        } catch (error) {
            console.error('Failed to load command history:', error);
            setOutput(prev => [...prev, { type: 'error', content: `Failed to load command history: ${error.message}` }]);
        }
    };

    const handleInputChange = (e) => {
        setInput(e.target.value);
    };

    const handleKeyDown = (e) => {
        if (e.key === 'ArrowUp') {
            e.preventDefault();
            if (historyIndex < commandHistory.length - 1) {
                setHistoryIndex(prevIndex => prevIndex + 1);
                setInput(commandHistory[historyIndex + 1]);
            }
        } else if (e.key === 'ArrowDown') {
            e.preventDefault();
            if (historyIndex > -1) {
                setHistoryIndex(prevIndex => prevIndex - 1);
                setInput(historyIndex - 1 < 0 ? '' : commandHistory[historyIndex - 1]);
            }
        } else if (e.key === 'Enter') {
            executeCommand();
        }
    };

    const executeCommand = async () => {
        if (!input.trim() || isRunning) return;

        setOutput(prev => [...prev, { type: 'input', content: input }]);
        setIsRunning(true);

        if (input.trim().toLowerCase() === 'clear') {
            setOutput([]);
            setIsRunning(false);
        } else {
            try {
                await window.go.main.App.ExecuteInteractiveCommand(activeProfile, input);
                await window.go.main.App.AddCommandToHistory(activeProfile, input);
                setCommandHistory(prev => [input, ...prev]);
                setHistoryIndex(-1);
            } catch (error) {
                console.error('Failed to execute command:', error);
                setOutput(prev => [...prev, { type: 'error', content: `Failed to execute command: ${error.message}` }]);
                setIsRunning(false);
            }
        }

        setInput('');
    };

    const stopCommand = async () => {
        try {
            setOutput(prev => [...prev, { type: 'info', content: 'Stopping command...' }]);
            await window.go.main.App.StopInteractiveCommand(activeProfile);
            setOutput(prev => [...prev, { type: 'info', content: 'Command stopped' }]);
            setIsRunning(false);
        } catch (error) {
            console.error('Failed to stop command:', error);
            setOutput(prev => [...prev, { type: 'error', content: `Failed to stop command: ${error.message}` }]);
        }
    };

    const clearTerminal = () => {
        setOutput([]);
    };

    const getOutputClassName = (type) => {
        switch (type) {
            case 'input':
                return 'text-blue-400';
            case 'error':
                return 'text-red-400';
            case 'stderr':
                return 'text-yellow-400';
            case 'info':
                return 'text-green-400';
            default:
                return 'text-gray-300';
        }
    };

    return (
        <div className="flex flex-col h-full bg-gray-900 text-gray-300 font-mono">
            <div ref={outputRef} className="flex-grow overflow-auto p-2">
                {output.map((item, index) => (
                    <div key={index} className={`mb-1 ${getOutputClassName(item.type)}`}>
                        {item.type === 'input' ? '$ ' : ''}{item.content}
                    </div>
                ))}
            </div>
            <div className="p-2 border-t border-gray-700 flex">
                <input
                    ref={inputRef}
                    type="text"
                    value={input}
                    onChange={handleInputChange}
                    onKeyDown={handleKeyDown}
                    className="flex-grow bg-gray-800 text-gray-300 px-2 py-1 mr-2 focus:outline-none rounded"
                    placeholder={isRunning ? "Command is running..." : "Enter command..."}
                    disabled={isRunning}
                />
                {isRunning ? (
                    <button
                        onClick={stopCommand}
                        className="bg-red-500 text-white px-4 py-2 rounded hover:bg-red-600 transition duration-200"
                    >
                        <FaStop className="mr-2 inline" /> Stop
                    </button>
                ) : (
                    <>
                        <button
                            onClick={executeCommand}
                            className="bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600 transition duration-200 mr-2"
                        >
                            <FaPlay className="mr-2 inline" /> Run
                        </button>
                        <button
                            onClick={clearTerminal}
                            className="bg-yellow-500 text-white px-4 py-2 rounded hover:bg-yellow-600 transition duration-200"
                        >
                            <FaTrash className="mr-2 inline" /> Clear
                        </button>
                    </>
                )}
            </div>
        </div>
    );
}