import React, { useState, useEffect, useRef } from 'react';
import { FaPlay, FaStop } from 'react-icons/fa';
import { FaFolder, FaFile, FaArrowUp, FaDownload, FaTrash, FaUpload } from 'react-icons/fa';


export function terminal2({ activeProfile }) {
    const [output, setOutput] = useState([]);
    const [input, setInput] = useState('');
    const [isRunning, setIsRunning] = useState(false);
    const [commandHistory, setCommandHistory] = useState([]);
    const [historyIndex, setHistoryIndex] = useState(-1);
    const outputRef = useRef(null);
    const inputRef = useRef(null);












}