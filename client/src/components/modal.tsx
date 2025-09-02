import React, { useEffect, useState } from 'react';
import { log } from '../utils';

export type ModalProperties = {
    children: React.JSX.Element;
    isOpen: boolean;
    close?: () => void;
    onResize?: (height: number) => void;
};

function Modal({ children, isOpen, close, onResize }: ModalProperties) {
    const [height, setHeight] = useState(0);

    const getModalHeight = () => {
        const elementList = document.getElementsByClassName('modal-inner');
        return elementList && elementList.length && elementList[0].clientHeight;
    };

    useEffect(() => {
        window.addEventListener('resize', () => {
            setHeight(getModalHeight());
        });
    }, []);

    useEffect(() => {
        onResize?.(height);
    }, [height]);

    const onClickOutsideClose = (e: object) => {
        if (
            'target' in e &&
            e.target &&
            typeof e.target === 'object' &&
            'className' in e.target &&
            e.target.className == 'modal'
        ) {
            close?.();
        }
    };

    if (!isOpen) return;

    setTimeout(() => {
        setHeight(getModalHeight());
    }, 10);

    return (
        <>
            <div className="modal" onClick={onClickOutsideClose}>
                <div className="modal-inner">{children}</div>
            </div>
        </>
    );
}

export default Modal;
