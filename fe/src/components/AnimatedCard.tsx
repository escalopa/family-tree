import React, { ReactNode } from 'react';
import { Card, CardProps } from '@mui/material';
import { motion } from 'framer-motion';

interface AnimatedCardProps extends CardProps {
  children: ReactNode;
  delay?: number;
}

const AnimatedCard: React.FC<AnimatedCardProps> = ({ children, delay = 0, ...props }) => {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{
        duration: 0.4,
        delay,
        ease: [0.4, 0.0, 0.2, 1],
      }}
      whileHover={{
        y: -4,
        transition: { duration: 0.2 },
      }}
    >
      <Card {...props}>{children}</Card>
    </motion.div>
  );
};

export default AnimatedCard;
