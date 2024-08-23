import React, { useState, useEffect, useRef } from 'react';
import { FaPlay, FaStop } from 'react-icons/fa';

export function Terminal({ activeProfile }) {
    const [output, setOutput] = useState([]);
    const [input, setInput] = useState('');
    const [isRunning, setIsRunning] = useState(false);
    const [commandHistory, setCommandHistory] = useState([]);
    const [historyIndex, setHistoryIndex] = useState(-1);
    const outputRef = useRef(null);
    const inputRef = useRef(null);

    useEffect(() => {
        const handleCommandOutput = (data) => {
            if (data.profile === activeProfile) {
                setOutput(prev => [...prev, { type: data.type, content: data.data }]);
                if (data.type === 'info' && (data.data.includes('Command finished') || data.data.includes('Command stopped'))) {
                    setIsRunning(false);
                }
                if (data.type === 'error') {
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

        try {
            await window.go.main.App.ExecuteInteractiveCommand(activeProfile, input);
            await window.go.main.App.AddCommandToHistory(activeProfile, input);
            setCommandHistory(prev => [input, ...prev]);
            setHistoryIndex(-1);
            await createSynonym(input);
        } catch (error) {
            setOutput(prev => [...prev, { type: 'error', content: `Failed to execute command: ${error.message}` }]);
            setIsRunning(false);
        }

        setInput('');
    };

    const stopCommand = async () => {
        try {
            await window.go.main.App.StopInteractiveCommand(activeProfile);
        } catch (error) {
            console.error('Failed to stop command:', error);
            setOutput(prev => [...prev, { type: 'error', content: `Failed to stop command: ${error.message}` }]);
            setIsRunning(false);
        }
    };

    const createSynonym = async (command) => {
        try {
            const synonym = await window.go.main.App.CreateSynonym(command);
            if (synonym) {
                setOutput(prev => [...prev, { type: 'info', content: `Created synonym: ${synonym} for command: ${command}` }]);
            }
        } catch (error) {
            console.error('Failed to create synonym:', error);
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
                    <button
                        onClick={executeCommand}
                        className="bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600 transition duration-200"
                    >
                        <FaPlay className="mr-2 inline" /> Run
                    </button>
                )}
            </div>
        </div>
    );
}